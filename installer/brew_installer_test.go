package installer

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/stretchr/testify/assert"
)

func newTestBrewInstaller(data *appconfig.InstallerData) *BrewInstaller {
	return &BrewInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Info: data,
	}
}

func TestBrewValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid: No tap specified (tap is optional)
	emptyData := &appconfig.InstallerData{
		Name: strPtr("test-brew-valid"),
		Type: appconfig.InstallerTypeBrew,
	}
	assertNoValidationErrors(t, newTestBrewInstaller(emptyData).Validate())

	// 🟢 Valid: Well-formed tap (contains slash, sufficient length)
	validData := &appconfig.InstallerData{
		Name: strPtr("test-brew-valid-tap"),
		Type: appconfig.InstallerTypeBrew,
		Opts: &map[string]any{"tap": "valid/tap"},
	}
	assertNoValidationErrors(t, newTestBrewInstaller(validData).Validate())

	// 🟢 Valid: Tap and cask used together
	tapCaskData := &appconfig.InstallerData{
		Name: strPtr("test-brew-tap-cask"),
		Type: appconfig.InstallerTypeBrew,
		Opts: &map[string]any{
			"tap":  "homebrew/cask-versions",
			"cask": true,
		},
	}
	assertNoValidationErrors(t, newTestBrewInstaller(tapCaskData).Validate())

	// 🔴 Invalid: Tap is present but malformed (missing slash or too short)
	invalidData := &appconfig.InstallerData{
		Name: strPtr("test-brew-invalid-tap"),
		Type: appconfig.InstallerTypeBrew,
		Opts: &map[string]any{"tap": "invalid-tap"},
	}
	assertHasValidationErrors(t, newTestBrewInstaller(invalidData).Validate())
}

func simulateBrewCheck(input string, exitCode int) (logs string, updateNeeded bool, finalErr error) {
	logBuf := &bytes.Buffer{}
	needsUpdate, parseErr := parseBrewOutdatedOutput(strings.NewReader(input), logBuf)

	// Treat only negative/128+ as actual errors (or change as needed)
	if exitCode < 0 || exitCode >= 128 {
		return logBuf.String(), false, fmt.Errorf("brew exited with error code %d", exitCode)
	}

	if parseErr != nil {
		return logBuf.String(), false, parseErr
	}

	// Exit code >0 means updates are available — trust that
	if exitCode > 0 {
		return logBuf.String(), true, nil
	}

	// Exit code 0: trust the parsed JSON
	return logBuf.String(), needsUpdate, nil
}

func TestBrewGetOpts(t *testing.T) {
	logger.InitLogger(false)

	// Test default opts (no options set)
	defaultData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypeBrew,
	}
	installer := newTestBrewInstaller(defaultData)
	opts := installer.GetOpts()
	if opts.Tap != nil {
		t.Errorf("expected Tap to be nil")
	}
	if opts.Cask != nil {
		t.Errorf("expected Cask to be nil")
	}
	if opts.Flags != nil {
		t.Errorf("expected Flags to be nil")
	}
	if opts.InstallFlags != nil {
		t.Errorf("expected InstallFlags to be nil")
	}
	if opts.UpdateFlags != nil {
		t.Errorf("expected UpdateFlags to be nil")
	}

	// Test with flags option
	flagsData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypeBrew,
		Opts: &map[string]any{
			"flags": "--verbose --debug",
		},
	}
	installerWithFlags := newTestBrewInstaller(flagsData)
	optsWithFlags := installerWithFlags.GetOpts()
	if optsWithFlags.Flags == nil || *optsWithFlags.Flags != "--verbose --debug" {
		t.Errorf("expected Flags to be '--verbose --debug'")
	}

	// Test with install_flags option
	installFlagsData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypeBrew,
		Opts: &map[string]any{
			"install_flags": "--force",
		},
	}
	installerWithInstallFlags := newTestBrewInstaller(installFlagsData)
	optsWithInstallFlags := installerWithInstallFlags.GetOpts()
	if optsWithInstallFlags.InstallFlags == nil || *optsWithInstallFlags.InstallFlags != "--force" {
		t.Errorf("expected InstallFlags to be '--force'")
	}

	// Test with update_flags option
	updateFlagsData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypeBrew,
		Opts: &map[string]any{
			"update_flags": "--dry-run",
		},
	}
	installerWithUpdateFlags := newTestBrewInstaller(updateFlagsData)
	optsWithUpdateFlags := installerWithUpdateFlags.GetOpts()
	if optsWithUpdateFlags.UpdateFlags == nil || *optsWithUpdateFlags.UpdateFlags != "--dry-run" {
		t.Errorf("expected UpdateFlags to be '--dry-run'")
	}

	// Test with all flags options combined
	allFlagsData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypeBrew,
		Opts: &map[string]any{
			"tap":           "homebrew/core",
			"cask":          true,
			"flags":         "--common",
			"install_flags": "--install-specific",
			"update_flags":  "--update-specific",
		},
	}
	installerWithAllFlags := newTestBrewInstaller(allFlagsData)
	optsWithAllFlags := installerWithAllFlags.GetOpts()
	if optsWithAllFlags.Tap == nil || *optsWithAllFlags.Tap != "homebrew/core" {
		t.Errorf("expected Tap to be 'homebrew/core'")
	}
	if optsWithAllFlags.Cask == nil || !*optsWithAllFlags.Cask {
		t.Errorf("expected Cask to be true")
	}
	if optsWithAllFlags.Flags == nil || *optsWithAllFlags.Flags != "--common" {
		t.Errorf("expected Flags to be '--common'")
	}
	if optsWithAllFlags.InstallFlags == nil || *optsWithAllFlags.InstallFlags != "--install-specific" {
		t.Errorf("expected InstallFlags to be '--install-specific'")
	}
	if optsWithAllFlags.UpdateFlags == nil || *optsWithAllFlags.UpdateFlags != "--update-specific" {
		t.Errorf("expected UpdateFlags to be '--update-specific'")
	}
}

func TestBrewNeedsUpdateWithExitCode(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		exitCode       int
		expectedLogs   string
		expectedUpdate bool
		expectErr      bool
	}{

		{
			name: "brew exit 1 (updates available)",
			input: `{
  "formulae": [],
  "casks": []
}`,
			exitCode:       1,
			expectedLogs:   "",
			expectedUpdate: true, // non-zero means updates
			expectErr:      false,
		},
		{
			name: "brew exit 0 (no updates)",
			input: `{
  "formulae": [],
  "casks": []
}`,
			exitCode:       0,
			expectedLogs:   "",
			expectedUpdate: false,
			expectErr:      false,
		},
		{
			name: "brew exit 1 with logs",
			input: `Auto-updating Homebrew...
{
  "formulae": [{ "name": "bash" }],
  "casks": []
}`,
			exitCode:       1,
			expectedLogs:   "Auto-updating Homebrew...\n",
			expectedUpdate: true,
			expectErr:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logs, update, err := simulateBrewCheck(tc.input, tc.exitCode)

			if tc.expectErr && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if update != tc.expectedUpdate {
				t.Errorf("unexpected update result: got %v, want %v", update, tc.expectedUpdate)
			}
			if logs != tc.expectedLogs {
				t.Errorf("unexpected logs:\nGot:\n%q\nWant:\n%q", logs, tc.expectedLogs)
			}
		})
	}
}

func TestBrewGetFullName(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns name without tap", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("vim"),
			Type: appconfig.InstallerTypeBrew,
		}
		installer := newTestBrewInstaller(data)
		assert.Equal(t, "vim", installer.GetFullName())
	})

	t.Run("returns tap/name with tap", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("sofmani"),
			Type: appconfig.InstallerTypeBrew,
			Opts: &map[string]any{
				"tap": "chenasraf/tap",
			},
		}
		installer := newTestBrewInstaller(data)
		assert.Equal(t, "chenasraf/tap/sofmani", installer.GetFullName())
	})
}

func TestBrewIsCask(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns false when cask is not set", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("vim"),
			Type: appconfig.InstallerTypeBrew,
		}
		installer := newTestBrewInstaller(data)
		assert.False(t, installer.IsCask())
	})

	t.Run("returns false when cask is false", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("firefox"),
			Type: appconfig.InstallerTypeBrew,
			Opts: &map[string]any{
				"cask": false,
			},
		}
		installer := newTestBrewInstaller(data)
		assert.False(t, installer.IsCask())
	})

	t.Run("returns true when cask is true", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("firefox"),
			Type: appconfig.InstallerTypeBrew,
			Opts: &map[string]any{
				"cask": true,
			},
		}
		installer := newTestBrewInstaller(data)
		assert.True(t, installer.IsCask())
	})
}

func TestBrewGetBinName(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns name when bin_name is not set", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("vim"),
			Type: appconfig.InstallerTypeBrew,
		}
		installer := newTestBrewInstaller(data)
		assert.Equal(t, "vim", installer.GetBinName())
	})

	t.Run("returns bin_name when set", func(t *testing.T) {
		binName := "nvim"
		data := &appconfig.InstallerData{
			Name:    strPtr("neovim"),
			Type:    appconfig.InstallerTypeBrew,
			BinName: &binName,
		}
		installer := newTestBrewInstaller(data)
		assert.Equal(t, "nvim", installer.GetBinName())
	})

	t.Run("returns name when bin_name is empty", func(t *testing.T) {
		binName := ""
		data := &appconfig.InstallerData{
			Name:    strPtr("vim"),
			Type:    appconfig.InstallerTypeBrew,
			BinName: &binName,
		}
		installer := newTestBrewInstaller(data)
		assert.Equal(t, "vim", installer.GetBinName())
	})
}

func TestBrewGetData(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns the installer data", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("vim"),
			Type: appconfig.InstallerTypeBrew,
		}
		installer := newTestBrewInstaller(data)
		result := installer.GetData()

		assert.Equal(t, data, result)
		assert.Equal(t, "vim", *result.Name)
	})
}

func TestNewBrewInstaller(t *testing.T) {
	logger.InitLogger(false)

	t.Run("creates installer with config and data", func(t *testing.T) {
		cfg := &appconfig.AppConfig{}
		data := &appconfig.InstallerData{
			Name: strPtr("vim"),
			Type: appconfig.InstallerTypeBrew,
		}
		installer := NewBrewInstaller(cfg, data)

		assert.NotNil(t, installer)
		assert.Equal(t, cfg, installer.Config)
		assert.Equal(t, data, installer.Info)
		assert.Equal(t, data, installer.Data)
	})
}

func TestBrewCheckIsInstalled(t *testing.T) {
	logger.InitLogger(false)

	t.Run("runs custom check when provided", func(t *testing.T) {
		checkCmd := "true"
		data := &appconfig.InstallerData{
			Name:           strPtr("test-brew"),
			Type:           appconfig.InstallerTypeBrew,
			CheckInstalled: &checkCmd,
		}
		installer := newTestBrewInstaller(data)
		result, err := installer.CheckIsInstalled()

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("runs custom check that fails", func(t *testing.T) {
		checkCmd := "false"
		data := &appconfig.InstallerData{
			Name:           strPtr("test-brew"),
			Type:           appconfig.InstallerTypeBrew,
			CheckInstalled: &checkCmd,
		}
		installer := newTestBrewInstaller(data)
		result, err := installer.CheckIsInstalled()

		assert.NoError(t, err)
		assert.False(t, result)
	})
}

func TestBrewCheckNeedsUpdate(t *testing.T) {
	logger.InitLogger(false)

	t.Run("runs custom check when provided", func(t *testing.T) {
		checkCmd := "true" // Returns exit code 0, meaning update available
		data := &appconfig.InstallerData{
			Name:           strPtr("test-brew"),
			Type:           appconfig.InstallerTypeBrew,
			CheckHasUpdate: &checkCmd,
		}
		installer := newTestBrewInstaller(data)
		result, err := installer.CheckNeedsUpdate()

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("custom check returns false when no update", func(t *testing.T) {
		checkCmd := "false" // Returns exit code 1, meaning no update
		data := &appconfig.InstallerData{
			Name:           strPtr("test-brew"),
			Type:           appconfig.InstallerTypeBrew,
			CheckHasUpdate: &checkCmd,
		}
		installer := newTestBrewInstaller(data)
		result, err := installer.CheckNeedsUpdate()

		assert.NoError(t, err)
		assert.False(t, result)
	})
}

func TestBrewGetOptsWrongTypes(t *testing.T) {
	logger.InitLogger(false)

	t.Run("handles wrong type values gracefully", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("vim"),
			Type: appconfig.InstallerTypeBrew,
			Opts: &map[string]any{
				"tap":           123,   // Wrong type
				"cask":          "yes", // Wrong type
				"flags":         true,  // Wrong type
				"install_flags": 456,   // Wrong type
				"update_flags":  false, // Wrong type
			},
		}
		installer := newTestBrewInstaller(data)
		opts := installer.GetOpts()

		// Should return nil when type assertion fails
		assert.Nil(t, opts.Tap)
		assert.Nil(t, opts.Cask)
		assert.Nil(t, opts.Flags)
		assert.Nil(t, opts.InstallFlags)
		assert.Nil(t, opts.UpdateFlags)
	})
}
