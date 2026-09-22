package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func newTestCargoInstaller(data *appconfig.InstallerData) *CargoInstaller {
	return &CargoInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config: nil,
		Info:   data,
	}
}

func TestCargoValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid cargo installer
	validData := &appconfig.InstallerData{
		Name: lo.ToPtr("ripgrep"),
		Type: appconfig.InstallerTypeCargo,
	}
	assertNoValidationErrors(t, newTestCargoInstaller(validData).Validate())

	// 🔴 Nil name
	nilNameData := &appconfig.InstallerData{
		Name: nil,
		Type: appconfig.InstallerTypeCargo,
	}
	assertValidationError(t, newTestCargoInstaller(nilNameData).Validate(), "name")
}

func TestCargoGetOpts(t *testing.T) {
	logger.InitLogger(false)

	// Test default opts (no options set)
	defaultData := &appconfig.InstallerData{
		Name: lo.ToPtr("ripgrep"),
		Type: appconfig.InstallerTypeCargo,
	}
	installer := newTestCargoInstaller(defaultData)
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
		Name: lo.ToPtr("ripgrep"),
		Type: appconfig.InstallerTypeCargo,
		Opts: &map[string]any{
			"flags": "--locked",
		},
	}
	installerWithFlags := newTestCargoInstaller(flagsData)
	optsWithFlags := installerWithFlags.GetOpts()
	if optsWithFlags.Flags == nil || *optsWithFlags.Flags != "--locked" {
		t.Errorf("expected Flags to be '--locked'")
	}

	// Test with install_flags option
	installFlagsData := &appconfig.InstallerData{
		Name: lo.ToPtr("ripgrep"),
		Type: appconfig.InstallerTypeCargo,
		Opts: &map[string]any{
			"install_flags": "--features pcre2",
		},
	}
	installerWithInstallFlags := newTestCargoInstaller(installFlagsData)
	optsWithInstallFlags := installerWithInstallFlags.GetOpts()
	if optsWithInstallFlags.InstallFlags == nil || *optsWithInstallFlags.InstallFlags != "--features pcre2" {
		t.Errorf("expected InstallFlags to be '--features pcre2'")
	}

	// Test with update_flags option
	updateFlagsData := &appconfig.InstallerData{
		Name: lo.ToPtr("ripgrep"),
		Type: appconfig.InstallerTypeCargo,
		Opts: &map[string]any{
			"update_flags": "--force",
		},
	}
	installerWithUpdateFlags := newTestCargoInstaller(updateFlagsData)
	optsWithUpdateFlags := installerWithUpdateFlags.GetOpts()
	if optsWithUpdateFlags.UpdateFlags == nil || *optsWithUpdateFlags.UpdateFlags != "--force" {
		t.Errorf("expected UpdateFlags to be '--force'")
	}

	// Test with all flags options combined
	allFlagsData := &appconfig.InstallerData{
		Name: lo.ToPtr("ripgrep"),
		Type: appconfig.InstallerTypeCargo,
		Opts: &map[string]any{
			"flags":         "--common",
			"install_flags": "--install-specific",
			"update_flags":  "--update-specific",
		},
	}
	installerWithAllFlags := newTestCargoInstaller(allFlagsData)
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

func TestCargoVersionPin(t *testing.T) {
	logger.InitLogger(false)

	unpinned := newTestCargoInstaller(&appconfig.InstallerData{
		Name: lo.ToPtr("ripgrep"),
		Type: appconfig.InstallerTypeCargo,
	})
	assert.Equal(t, "", unpinned.GetPinnedVersion())
	assert.Nil(t, unpinned.GetVersionArgs())

	pinned := newTestCargoInstaller(&appconfig.InstallerData{
		Name: lo.ToPtr("ripgrep"),
		Type: appconfig.InstallerTypeCargo,
		Opts: &map[string]any{"version": "14.1.0"},
	})
	assert.Equal(t, "14.1.0", pinned.GetPinnedVersion())
	assert.Equal(t, []string{"--version", "14.1.0"}, pinned.GetVersionArgs())
}

func TestCargoGetBinName(t *testing.T) {
	logger.InitLogger(false)

	// Default: uses installer name
	defaultData := &appconfig.InstallerData{
		Name: lo.ToPtr("ripgrep"),
		Type: appconfig.InstallerTypeCargo,
	}
	installer := newTestCargoInstaller(defaultData)
	if installer.GetBinName() != "ripgrep" {
		t.Errorf("expected bin name to be 'ripgrep', got '%s'", installer.GetBinName())
	}

	// Override: uses bin_name
	binNameData := &appconfig.InstallerData{
		Name:    lo.ToPtr("ripgrep"),
		Type:    appconfig.InstallerTypeCargo,
		BinName: lo.ToPtr("rg"),
	}
	installerWithBinName := newTestCargoInstaller(binNameData)
	if installerWithBinName.GetBinName() != "rg" {
		t.Errorf("expected bin name to be 'rg', got '%s'", installerWithBinName.GetBinName())
	}
}

func TestCargoGetData(t *testing.T) {
	logger.InitLogger(false)

	data := &appconfig.InstallerData{
		Name: lo.ToPtr("ripgrep"),
		Type: appconfig.InstallerTypeCargo,
	}
	installer := newTestCargoInstaller(data)
	if installer.GetData() != data {
		t.Errorf("expected GetData to return the same data pointer")
	}
}
