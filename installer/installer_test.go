package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
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

func TestCheckIsInstalled_UsesBinName(t *testing.T) {
	logger.InitLogger(false)

	// which-based installers: bin_name is "ls" (exists), name is something that doesn't exist.
	// If CheckIsInstalled respects bin_name, it returns true.
	whichBasedCases := []struct {
		name          string
		newInstaller  func(*appconfig.InstallerData) IInstaller
		installerType appconfig.InstallerType
	}{
		{"brew", func(d *appconfig.InstallerData) IInstaller { return newTestBrewInstaller(d) }, appconfig.InstallerTypeBrew},
		{"shell", func(d *appconfig.InstallerData) IInstaller { return newTestShellInstaller(d) }, appconfig.InstallerTypeShell},
		{"npm", func(d *appconfig.InstallerData) IInstaller { return newTestNpmInstaller(d) }, appconfig.InstallerTypeNpm},
		{"apt", func(d *appconfig.InstallerData) IInstaller { return newTestAptInstaller(d) }, appconfig.InstallerTypeApt},
		{"pipx", func(d *appconfig.InstallerData) IInstaller { return newTestPipxInstaller(d) }, appconfig.InstallerTypePipx},
		{"cargo", func(d *appconfig.InstallerData) IInstaller { return newTestCargoInstaller(d) }, appconfig.InstallerTypeCargo},
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
