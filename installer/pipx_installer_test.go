package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
)

func newTestPipxInstaller(data *appconfig.InstallerData) *PipxInstaller {
	return &PipxInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config: nil,
		Info:   data,
	}
}

func TestPipxValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid pipx installer
	validData := &appconfig.InstallerData{
		Name: strPtr("some-pipx-package"),
		Type: appconfig.InstallerTypePipx,
	}
	assertNoValidationErrors(t, newTestPipxInstaller(validData).Validate())

	// 🔴 Optional: test nil name if BaseValidate handles it
	nilNameData := &appconfig.InstallerData{
		Name: nil,
		Type: appconfig.InstallerTypePipx,
	}
	assertValidationError(t, newTestPipxInstaller(nilNameData).Validate(), "name")
}

func TestPipxGetOpts(t *testing.T) {
	logger.InitLogger(false)

	// Test default opts (no options set)
	defaultData := &appconfig.InstallerData{
		Name: strPtr("black"),
		Type: appconfig.InstallerTypePipx,
	}
	installer := newTestPipxInstaller(defaultData)
	opts := installer.GetOpts()
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
		Name: strPtr("black"),
		Type: appconfig.InstallerTypePipx,
		Opts: &map[string]any{
			"flags": "--verbose",
		},
	}
	installerWithFlags := newTestPipxInstaller(flagsData)
	optsWithFlags := installerWithFlags.GetOpts()
	if optsWithFlags.Flags == nil || *optsWithFlags.Flags != "--verbose" {
		t.Errorf("expected Flags to be '--verbose'")
	}

	// Test with install_flags option
	installFlagsData := &appconfig.InstallerData{
		Name: strPtr("black"),
		Type: appconfig.InstallerTypePipx,
		Opts: &map[string]any{
			"install_flags": "--python python3.11",
		},
	}
	installerWithInstallFlags := newTestPipxInstaller(installFlagsData)
	optsWithInstallFlags := installerWithInstallFlags.GetOpts()
	if optsWithInstallFlags.InstallFlags == nil || *optsWithInstallFlags.InstallFlags != "--python python3.11" {
		t.Errorf("expected InstallFlags to be '--python python3.11'")
	}

	// Test with update_flags option
	updateFlagsData := &appconfig.InstallerData{
		Name: strPtr("black"),
		Type: appconfig.InstallerTypePipx,
		Opts: &map[string]any{
			"update_flags": "--force",
		},
	}
	installerWithUpdateFlags := newTestPipxInstaller(updateFlagsData)
	optsWithUpdateFlags := installerWithUpdateFlags.GetOpts()
	if optsWithUpdateFlags.UpdateFlags == nil || *optsWithUpdateFlags.UpdateFlags != "--force" {
		t.Errorf("expected UpdateFlags to be '--force'")
	}

	// Test with all flags options combined
	allFlagsData := &appconfig.InstallerData{
		Name: strPtr("black"),
		Type: appconfig.InstallerTypePipx,
		Opts: &map[string]any{
			"flags":         "--common",
			"install_flags": "--install-specific",
			"update_flags":  "--update-specific",
		},
	}
	installerWithAllFlags := newTestPipxInstaller(allFlagsData)
	optsWithAllFlags := installerWithAllFlags.GetOpts()
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
