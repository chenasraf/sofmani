package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
)

func newAptInstaller(data *appconfig.InstallerData) *AptInstaller {
	return &AptInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config: nil,
		Info:   data,
	}
}

func TestAptValidation(t *testing.T) {
	logger.InitLogger(false)
	aptInstaller := newAptInstaller(
		&appconfig.InstallerData{
			Name: strPtr("test-apt"),
			Type: appconfig.InstallerTypeApt,
		},
	)
	assertNoValidationErrors(t, aptInstaller.Validate())
}

func TestAptGetOpts(t *testing.T) {
	logger.InitLogger(false)

	// Test default opts (no options set)
	defaultData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypeApt,
	}
	installer := newAptInstaller(defaultData)
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
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypeApt,
		Opts: &map[string]any{
			"flags": "-y --no-install-recommends",
		},
	}
	installerWithFlags := newAptInstaller(flagsData)
	optsWithFlags := installerWithFlags.GetOpts()
	if optsWithFlags.Flags == nil || *optsWithFlags.Flags != "-y --no-install-recommends" {
		t.Errorf("expected Flags to be '-y --no-install-recommends'")
	}

	// Test with install_flags option
	installFlagsData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypeApt,
		Opts: &map[string]any{
			"install_flags": "--no-install-recommends",
		},
	}
	installerWithInstallFlags := newAptInstaller(installFlagsData)
	optsWithInstallFlags := installerWithInstallFlags.GetOpts()
	if optsWithInstallFlags.InstallFlags == nil || *optsWithInstallFlags.InstallFlags != "--no-install-recommends" {
		t.Errorf("expected InstallFlags to be '--no-install-recommends'")
	}

	// Test with update_flags option
	updateFlagsData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypeApt,
		Opts: &map[string]any{
			"update_flags": "--only-upgrade",
		},
	}
	installerWithUpdateFlags := newAptInstaller(updateFlagsData)
	optsWithUpdateFlags := installerWithUpdateFlags.GetOpts()
	if optsWithUpdateFlags.UpdateFlags == nil || *optsWithUpdateFlags.UpdateFlags != "--only-upgrade" {
		t.Errorf("expected UpdateFlags to be '--only-upgrade'")
	}

	// Test with all flags options combined
	allFlagsData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypeApt,
		Opts: &map[string]any{
			"flags":         "--common",
			"install_flags": "--install-specific",
			"update_flags":  "--update-specific",
		},
	}
	installerWithAllFlags := newAptInstaller(allFlagsData)
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
