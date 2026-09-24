package installer

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/summary"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestGetInstaller(t *testing.T) {
	config := &appconfig.AppConfig{}
	logger.InitLogger(false)
	installer := &appconfig.InstallerData{Type: appconfig.InstallerTypeBrew}
	inst, err := GetInstaller(config, installer)
	assert.NoError(t, err)
	assert.NotNil(t, inst)
}

func TestInstallerWithDefaults(t *testing.T) {
	opts := map[string]any{"key": "value"}
	defaults := &appconfig.AppConfigDefaults{
		Type: &map[appconfig.InstallerType]appconfig.InstallerData{
			appconfig.InstallerTypeBrew: {Opts: &opts},
		},
	}
	installer := &appconfig.InstallerData{Type: appconfig.InstallerTypeBrew, Opts: &map[string]any{}}
	result := InstallerWithDefaults(installer, appconfig.InstallerTypeBrew, defaults)
	assert.Equal(t, "value", (*result.Opts)["key"])
}

func TestRunInstaller(t *testing.T) {
	config := &appconfig.AppConfig{}
	mockInstaller := &MockInstaller{
		data:        &appconfig.InstallerData{Name: lo.ToPtr("test"), Type: appconfig.InstallerTypeBrew},
		isInstalled: false,
	}
	result, err := RunInstaller(config, mockInstaller)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestRunInstallerAllowFailure(t *testing.T) {
	logger.InitLogger(false)
	config := &appconfig.AppConfig{}
	failure := errors.New("install blew up")

	// Without allow_failure the error reaches the caller, which stops the run.
	strict := &MockInstaller{
		data:         &appconfig.InstallerData{Name: lo.ToPtr("strict"), Type: appconfig.InstallerTypeBrew},
		installError: failure,
	}
	result, err := RunInstaller(config, strict)
	assert.ErrorIs(t, err, failure)
	assert.Nil(t, result)

	// With allow_failure the error is reported and swallowed, leaving no result to summarize.
	lenient := &MockInstaller{
		data: &appconfig.InstallerData{
			Name:         lo.ToPtr("lenient"),
			Type:         appconfig.InstallerTypeBrew,
			AllowFailure: lo.ToPtr(true),
		},
		installError: failure,
	}
	result, err = RunInstaller(config, lenient)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestRunInstallerConfirm(t *testing.T) {
	logger.InitLogger(false)
	config := &appconfig.AppConfig{CheckUpdates: lo.ToPtr(true)}

	// answerWith replaces the prompt for the duration of a test.
	answerWith := func(t *testing.T, answer bool) *[]string {
		t.Helper()
		asked := []string{}
		original := confirmPrompt
		confirmPrompt = func(question string) bool {
			asked = append(asked, question)
			return answer
		}
		t.Cleanup(func() { confirmPrompt = original })
		return &asked
	}

	t.Run("install is skipped when declined", func(t *testing.T) {
		asked := answerWith(t, false)
		inst := &MockInstaller{
			data: &appconfig.InstallerData{
				Name:           lo.ToPtr("declined"),
				Type:           appconfig.InstallerTypeBrew,
				ConfirmInstall: lo.ToPtr(true),
			},
		}
		result, err := RunInstaller(config, inst)
		assert.NoError(t, err)
		assert.Equal(t, summary.ActionSkipped, result.Action)
		assert.Zero(t, inst.installCalls)
		assert.Equal(t, []string{"Install brew: declined?"}, *asked)
	})

	t.Run("install proceeds when accepted", func(t *testing.T) {
		answerWith(t, true)
		inst := &MockInstaller{
			data: &appconfig.InstallerData{
				Name:           lo.ToPtr("accepted"),
				Type:           appconfig.InstallerTypeBrew,
				ConfirmInstall: lo.ToPtr(true),
			},
		}
		result, err := RunInstaller(config, inst)
		assert.NoError(t, err)
		assert.Equal(t, summary.ActionInstalled, result.Action)
		assert.Equal(t, 1, inst.installCalls)
	})

	t.Run("update is skipped when declined", func(t *testing.T) {
		asked := answerWith(t, false)
		inst := &MockInstaller{
			data: &appconfig.InstallerData{
				Name:          lo.ToPtr("declined-update"),
				Type:          appconfig.InstallerTypeNpm,
				ConfirmUpdate: lo.ToPtr(true),
			},
			isInstalled: true,
			needsUpdate: true,
		}
		result, err := RunInstaller(config, inst)
		assert.NoError(t, err)
		assert.Equal(t, summary.ActionSkipped, result.Action)
		assert.Zero(t, inst.updateCalls)
		assert.Equal(t, []string{"Update npm: declined-update?"}, *asked)
	})

	t.Run("confirm_install leaves an update alone", func(t *testing.T) {
		asked := answerWith(t, false)
		inst := &MockInstaller{
			data: &appconfig.InstallerData{
				Name:           lo.ToPtr("install-only"),
				Type:           appconfig.InstallerTypeBrew,
				ConfirmInstall: lo.ToPtr(true),
			},
			isInstalled: true,
			needsUpdate: true,
		}
		result, err := RunInstaller(config, inst)
		assert.NoError(t, err)
		assert.Equal(t, summary.ActionUpgraded, result.Action)
		assert.Equal(t, 1, inst.updateCalls)
		assert.Empty(t, *asked)
	})

	t.Run("nothing is asked without the flags", func(t *testing.T) {
		asked := answerWith(t, false)
		inst := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("quiet"), Type: appconfig.InstallerTypeBrew},
		}
		result, err := RunInstaller(config, inst)
		assert.NoError(t, err)
		assert.Equal(t, summary.ActionInstalled, result.Action)
		assert.Empty(t, *asked)
	})
}

// pinnedMockInstaller is a MockInstaller pinned to a version.
type pinnedMockInstaller struct {
	*MockInstaller
	version string
}

// GetPinnedVersion implements IVersionPinned.
func (m *pinnedMockInstaller) GetPinnedVersion() string {
	return m.version
}

func TestRunInstallerRecordsPinnedVersion(t *testing.T) {
	logger.InitLogger(false)
	config := &appconfig.AppConfig{}
	name := "test-pin-run-installer"
	mockInstaller := &pinnedMockInstaller{
		MockInstaller: &MockInstaller{
			data:        &appconfig.InstallerData{Name: lo.ToPtr(name), Type: appconfig.InstallerTypeNpm},
			isInstalled: false,
		},
		version: "1.2.3",
	}
	t.Cleanup(func() {
		if file, err := pinnedVersionCacheFile(name); err == nil {
			_ = os.Remove(file)
		}
	})

	result, err := RunInstaller(config, mockInstaller)
	assert.NoError(t, err)
	assert.Equal(t, summary.ActionInstalled, result.Action)
	assert.Equal(t, "1.2.3", ReadInstalledVersion(name))
}

func TestCheckIsInstalled_UsesBinName(t *testing.T) {
	logger.InitLogger(false)

	// which-based installers: bin_name is "ls" (exists), name is something that doesn't exist.
	// If CheckIsInstalled respects bin_name, it returns true.
	whichBasedCases := []struct {
		name          string
		newInstaller  func(*appconfig.InstallerData) IInstaller
		installerType appconfig.InstallerType
	}{
		{"shell", func(d *appconfig.InstallerData) IInstaller { return newTestShellInstaller(d) }, appconfig.InstallerTypeShell},
		{"npm", func(d *appconfig.InstallerData) IInstaller { return newTestNpmInstaller(d) }, appconfig.InstallerTypeNpm},
		{"apt", func(d *appconfig.InstallerData) IInstaller { return newTestAptInstaller(d) }, appconfig.InstallerTypeApt},
		{"pipx", func(d *appconfig.InstallerData) IInstaller { return newTestPipxInstaller(d) }, appconfig.InstallerTypePipx},
		{"cargo", func(d *appconfig.InstallerData) IInstaller { return newTestCargoInstaller(d) }, appconfig.InstallerTypeCargo},
		{"go", func(d *appconfig.InstallerData) IInstaller { return newTestGoInstaller(d) }, appconfig.InstallerTypeGo},
		{"group", func(d *appconfig.InstallerData) IInstaller { return newTestGroupInstaller(d) }, appconfig.InstallerTypeGroup},
	}

	for _, tc := range whichBasedCases {
		t.Run(tc.name+"_uses_bin_name", func(t *testing.T) {
			data := &appconfig.InstallerData{
				Name:    lo.ToPtr("nonexistent-bin-xyz-12345"),
				BinName: lo.ToPtr("ls"),
				Type:    tc.installerType,
			}
			installer := tc.newInstaller(data)
			installed, err := installer.CheckIsInstalled()
			assert.NoError(t, err)
			assert.True(t, installed, "%s should use bin_name (ls) for install check, not name", tc.name)
		})

		t.Run(tc.name+"_falls_back_to_name", func(t *testing.T) {
			data := &appconfig.InstallerData{
				Name: lo.ToPtr("ls"),
				Type: tc.installerType,
			}
			installer := tc.newInstaller(data)
			installed, err := installer.CheckIsInstalled()
			assert.NoError(t, err)
			assert.True(t, installed, "%s should fall back to name (ls) when bin_name is not set", tc.name)
		})
	}

	// github-release: uses file path check with bin_name
	t.Run("github-release_uses_bin_name", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "sofmani-install-test")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		// Create binary with bin_name, not the installer name
		err = os.WriteFile(filepath.Join(tmpDir, "cospend"), []byte("fake"), 0755)
		assert.NoError(t, err)

		data := &appconfig.InstallerData{
			Name:    lo.ToPtr("cospend-cli"),
			BinName: lo.ToPtr("cospend"),
			Type:    appconfig.InstallerTypeGitHubRelease,
			Opts:    &map[string]any{"destination": tmpDir},
		}
		installer := newTestGitHubReleaseInstaller(data)
		installed, err := installer.CheckIsInstalled()
		assert.NoError(t, err)
		assert.True(t, installed, "github-release should use bin_name for install check path")
	})

	t.Run("github-release_falls_back_to_name", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "sofmani-install-test")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		err = os.WriteFile(filepath.Join(tmpDir, "myapp"), []byte("fake"), 0755)
		assert.NoError(t, err)

		data := &appconfig.InstallerData{
			Name: lo.ToPtr("myapp"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{"destination": tmpDir},
		}
		installer := newTestGitHubReleaseInstaller(data)
		installed, err := installer.CheckIsInstalled()
		assert.NoError(t, err)
		assert.True(t, installed, "github-release should fall back to name when bin_name is not set")
	})
}
