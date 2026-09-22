package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
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
		Name: lo.ToPtr("some-npm-package"),
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

func TestNpmVersionPin(t *testing.T) {
	logger.InitLogger(false)

	unpinned := newTestNpmInstaller(&appconfig.InstallerData{
		Name: lo.ToPtr("prettier"),
		Type: appconfig.InstallerTypeNpm,
	})
	assert.Equal(t, "", unpinned.GetPinnedVersion())
	assert.Equal(t, "prettier", unpinned.GetPackageSpec())

	pinned := newTestNpmInstaller(&appconfig.InstallerData{
		Name: lo.ToPtr("prettier"),
		Type: appconfig.InstallerTypeNpm,
		Opts: &map[string]any{"version": "3.3.3"},
	})
	assert.Equal(t, "3.3.3", pinned.GetPinnedVersion())
	assert.Equal(t, "prettier@3.3.3", pinned.GetPackageSpec())
	assert.Equal(t, "prettier", pinned.GetBinName())

	// A version written onto the name pins just the same.
	inline := newTestNpmInstaller(&appconfig.InstallerData{
		Name: lo.ToPtr("prettier@3.3.3"),
		Type: appconfig.InstallerTypeNpm,
	})
	assert.Equal(t, "3.3.3", inline.GetPinnedVersion())
	assert.Equal(t, "prettier@3.3.3", inline.GetPackageSpec())
	assert.Equal(t, "prettier", inline.GetBinName())

	// The `@` opening a scoped package is not a version.
	scoped := newTestNpmInstaller(&appconfig.InstallerData{
		Name: lo.ToPtr("@vue/cli"),
		Type: appconfig.InstallerTypeNpm,
		Opts: &map[string]any{"version": "5.0.8"},
	})
	assert.Equal(t, "5.0.8", scoped.GetPinnedVersion())
	assert.Equal(t, "@vue/cli@5.0.8", scoped.GetPackageSpec())

	scopedUnpinned := newTestNpmInstaller(&appconfig.InstallerData{
		Name: lo.ToPtr("@vue/cli"),
		Type: appconfig.InstallerTypeNpm,
	})
	assert.Equal(t, "", scopedUnpinned.GetPinnedVersion())
	assert.Equal(t, "@vue/cli", scopedUnpinned.GetBinName())
}

func TestNpmGetOpts(t *testing.T) {
	logger.InitLogger(false)

	// Test default opts (no options set)
	defaultData := &appconfig.InstallerData{
		Name: lo.ToPtr("prettier"),
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
		Name: lo.ToPtr("prettier"),
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
		Name: lo.ToPtr("prettier"),
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
		Name: lo.ToPtr("prettier"),
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
		Name: lo.ToPtr("prettier"),
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
