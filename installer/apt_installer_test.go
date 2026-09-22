package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func newTestAptInstaller(data *appconfig.InstallerData) *AptInstaller {
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
	aptInstaller := newTestAptInstaller(
		&appconfig.InstallerData{
			Name: lo.ToPtr("test-apt"),
			Type: appconfig.InstallerTypeApt,
		},
	)
	assertNoValidationErrors(t, aptInstaller.Validate())
}

func TestAptVersionPin(t *testing.T) {
	logger.InitLogger(false)

	unpinned := newTestAptInstaller(&appconfig.InstallerData{
		Name: lo.ToPtr("ripgrep"),
		Type: appconfig.InstallerTypeApt,
	})
	assert.Equal(t, "", unpinned.GetPinnedVersion())
	assert.Equal(t, "ripgrep", unpinned.GetPackageSpec())

	pinned := newTestAptInstaller(&appconfig.InstallerData{
		Name: lo.ToPtr("ripgrep"),
		Type: appconfig.InstallerTypeApt,
		Opts: &map[string]any{"version": "13.0.0-2"},
	})
	assert.Equal(t, "13.0.0-2", pinned.GetPinnedVersion())
	assert.Equal(t, "ripgrep=13.0.0-2", pinned.GetPackageSpec())
	assert.Equal(t, "ripgrep", pinned.GetBinName())

	// A version written onto the name pins just the same.
	inline := newTestAptInstaller(&appconfig.InstallerData{
		Name: lo.ToPtr("ripgrep=13.0.0-2"),
		Type: appconfig.InstallerTypeApt,
	})
	assert.Equal(t, "13.0.0-2", inline.GetPinnedVersion())
	assert.Equal(t, "ripgrep=13.0.0-2", inline.GetPackageSpec())
	assert.Equal(t, "ripgrep", inline.GetBinName())
}

func TestAptInstallVerb(t *testing.T) {
	logger.InitLogger(false)

	apt := newTestAptInstaller(&appconfig.InstallerData{
		Name: lo.ToPtr("ripgrep"),
		Type: appconfig.InstallerTypeApt,
	})
	apt.PackageManager = PackageManagerApt
	assert.Equal(t, "install", apt.installVerb())

	apk := newTestAptInstaller(&appconfig.InstallerData{
		Name: lo.ToPtr("ripgrep"),
		Type: appconfig.InstallerTypeApk,
	})
	apk.PackageManager = PackageManagerApk
	assert.Equal(t, "add", apk.installVerb())
}

func TestAptGetOpts(t *testing.T) {
	logger.InitLogger(false)

	// Test default opts (no options set)
	defaultData := &appconfig.InstallerData{
		Name: lo.ToPtr("vim"),
		Type: appconfig.InstallerTypeApt,
	}
	installer := newTestAptInstaller(defaultData)
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
		Name: lo.ToPtr("vim"),
		Type: appconfig.InstallerTypeApt,
		Opts: &map[string]any{
			"flags": "-y --no-install-recommends",
		},
	}
	installerWithFlags := newTestAptInstaller(flagsData)
	optsWithFlags := installerWithFlags.GetOpts()
	if optsWithFlags.Flags == nil || *optsWithFlags.Flags != "-y --no-install-recommends" {
		t.Errorf("expected Flags to be '-y --no-install-recommends'")
	}

	// Test with install_flags option
	installFlagsData := &appconfig.InstallerData{
		Name: lo.ToPtr("vim"),
		Type: appconfig.InstallerTypeApt,
		Opts: &map[string]any{
			"install_flags": "--no-install-recommends",
		},
	}
	installerWithInstallFlags := newTestAptInstaller(installFlagsData)
	optsWithInstallFlags := installerWithInstallFlags.GetOpts()
	if optsWithInstallFlags.InstallFlags == nil || *optsWithInstallFlags.InstallFlags != "--no-install-recommends" {
		t.Errorf("expected InstallFlags to be '--no-install-recommends'")
	}

	// Test with update_flags option
	updateFlagsData := &appconfig.InstallerData{
		Name: lo.ToPtr("vim"),
		Type: appconfig.InstallerTypeApt,
		Opts: &map[string]any{
			"update_flags": "--only-upgrade",
		},
	}
	installerWithUpdateFlags := newTestAptInstaller(updateFlagsData)
	optsWithUpdateFlags := installerWithUpdateFlags.GetOpts()
	if optsWithUpdateFlags.UpdateFlags == nil || *optsWithUpdateFlags.UpdateFlags != "--only-upgrade" {
		t.Errorf("expected UpdateFlags to be '--only-upgrade'")
	}

	// Test with all flags options combined
	allFlagsData := &appconfig.InstallerData{
		Name: lo.ToPtr("vim"),
		Type: appconfig.InstallerTypeApt,
		Opts: &map[string]any{
			"flags":         "--common",
			"install_flags": "--install-specific",
			"update_flags":  "--update-specific",
		},
	}
	installerWithAllFlags := newTestAptInstaller(allFlagsData)
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
