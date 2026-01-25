package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
)

func newTestNpmInstaller(data *appconfig.InstallerData) *NpmInstaller {
	return &NpmInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config:         nil,
		PackageManager: PackageManagerNpm,
		Info:           data,
	}
}

func TestNpmValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid npm installer
	validData := &appconfig.InstallerData{
		Name: strPtr("some-npm-package"),
		Type: appconfig.InstallerTypeNpm,
	}
	assertNoValidationErrors(t, newTestNpmInstaller(validData).Validate())

	// 🔴 Edge case: nil name (will panic or fail in BaseValidate if implemented to check it)
	nilNameData := &appconfig.InstallerData{
		Name: nil,
		Type: appconfig.InstallerTypeNpm,
	}
	assertValidationError(t, newTestNpmInstaller(nilNameData).Validate(), "name")
}

func TestNpmGetOpts(t *testing.T) {
	logger.InitLogger(false)

	// Test default opts (no options set)
	defaultData := &appconfig.InstallerData{
		Name: strPtr("prettier"),
		Type: appconfig.InstallerTypeNpm,
	}
	installer := newTestNpmInstaller(defaultData)
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
		Name: strPtr("prettier"),
		Type: appconfig.InstallerTypeNpm,
		Opts: &map[string]any{
			"flags": "--legacy-peer-deps",
		},
	}
	installerWithFlags := newTestNpmInstaller(flagsData)
	optsWithFlags := installerWithFlags.GetOpts()
	if optsWithFlags.Flags == nil || *optsWithFlags.Flags != "--legacy-peer-deps" {
		t.Errorf("expected Flags to be '--legacy-peer-deps'")
	}

	// Test with install_flags option
	installFlagsData := &appconfig.InstallerData{
		Name: strPtr("prettier"),
		Type: appconfig.InstallerTypeNpm,
		Opts: &map[string]any{
			"install_flags": "--save-exact",
		},
	}
	installerWithInstallFlags := newTestNpmInstaller(installFlagsData)
	optsWithInstallFlags := installerWithInstallFlags.GetOpts()
	if optsWithInstallFlags.InstallFlags == nil || *optsWithInstallFlags.InstallFlags != "--save-exact" {
		t.Errorf("expected InstallFlags to be '--save-exact'")
	}

	// Test with update_flags option
	updateFlagsData := &appconfig.InstallerData{
		Name: strPtr("prettier"),
		Type: appconfig.InstallerTypeNpm,
		Opts: &map[string]any{
			"update_flags": "--force",
		},
	}
	installerWithUpdateFlags := newTestNpmInstaller(updateFlagsData)
	optsWithUpdateFlags := installerWithUpdateFlags.GetOpts()
	if optsWithUpdateFlags.UpdateFlags == nil || *optsWithUpdateFlags.UpdateFlags != "--force" {
		t.Errorf("expected UpdateFlags to be '--force'")
	}

	// Test with all flags options combined
	allFlagsData := &appconfig.InstallerData{
		Name: strPtr("prettier"),
		Type: appconfig.InstallerTypeNpm,
		Opts: &map[string]any{
			"flags":         "--common",
			"install_flags": "--install-specific",
			"update_flags":  "--update-specific",
		},
	}
	installerWithAllFlags := newTestNpmInstaller(allFlagsData)
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
