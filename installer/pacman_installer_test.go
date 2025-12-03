package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
)

func newTestPacmanInstaller(data *appconfig.InstallerData) *PacmanInstaller {
	return &PacmanInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config:         nil,
		PackageManager: PackageManagerPacman,
		Info:           data,
	}
}

func newTestYayInstaller(data *appconfig.InstallerData) *PacmanInstaller {
	return &PacmanInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config:         nil,
		PackageManager: PackageManagerYay,
		Info:           data,
	}
}

func TestPacmanValidation(t *testing.T) {
	logger.InitLogger(false)

	// Valid pacman installer
	validPacmanData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypePacman,
	}
	assertNoValidationErrors(t, newTestPacmanInstaller(validPacmanData).Validate())

	// Valid yay installer
	validYayData := &appconfig.InstallerData{
		Name: strPtr("visual-studio-code-bin"),
		Type: appconfig.InstallerTypeYay,
	}
	assertNoValidationErrors(t, newTestYayInstaller(validYayData).Validate())

	// Invalid: nil name
	nilNameData := &appconfig.InstallerData{
		Name: nil,
		Type: appconfig.InstallerTypePacman,
	}
	assertValidationError(t, newTestPacmanInstaller(nilNameData).Validate(), "name")
}

func TestPacmanGetBinName(t *testing.T) {
	logger.InitLogger(false)

	// Test default bin name (uses package name)
	defaultBinData := &appconfig.InstallerData{
		Name: strPtr("neovim"),
		Type: appconfig.InstallerTypePacman,
	}
	installer := newTestPacmanInstaller(defaultBinData)
	if installer.GetBinName() != "neovim" {
		t.Errorf("expected bin name 'neovim', got '%s'", installer.GetBinName())
	}

	// Test custom bin name
	customBinName := "nvim"
	customBinData := &appconfig.InstallerData{
		Name:    strPtr("neovim"),
		Type:    appconfig.InstallerTypePacman,
		BinName: &customBinName,
	}
	installerWithCustomBin := newTestPacmanInstaller(customBinData)
	if installerWithCustomBin.GetBinName() != "nvim" {
		t.Errorf("expected bin name 'nvim', got '%s'", installerWithCustomBin.GetBinName())
	}
}

func TestPacmanGetOpts(t *testing.T) {
	logger.InitLogger(false)

	// Test default opts (no options set)
	defaultData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypePacman,
	}
	installer := newTestPacmanInstaller(defaultData)
	opts := installer.GetOpts()
	if opts.Needed != nil {
		t.Errorf("expected Needed to be nil, got %v", *opts.Needed)
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

	// Test with needed option set to true
	neededData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypePacman,
		Opts: &map[string]any{
			"needed": true,
		},
	}
	installerWithNeeded := newTestPacmanInstaller(neededData)
	optsWithNeeded := installerWithNeeded.GetOpts()
	if optsWithNeeded.Needed == nil || !*optsWithNeeded.Needed {
		t.Errorf("expected Needed to be true")
	}

	// Test with needed option set to false
	notNeededData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypePacman,
		Opts: &map[string]any{
			"needed": false,
		},
	}
	installerNotNeeded := newTestPacmanInstaller(notNeededData)
	optsNotNeeded := installerNotNeeded.GetOpts()
	if optsNotNeeded.Needed == nil || *optsNotNeeded.Needed {
		t.Errorf("expected Needed to be false")
	}

	// Test with flags option
	flagsData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypePacman,
		Opts: &map[string]any{
			"flags": "--asdeps --overwrite '*'",
		},
	}
	installerWithFlags := newTestPacmanInstaller(flagsData)
	optsWithFlags := installerWithFlags.GetOpts()
	if optsWithFlags.Flags == nil || *optsWithFlags.Flags != "--asdeps --overwrite '*'" {
		t.Errorf("expected Flags to be '--asdeps --overwrite '*''")
	}

	// Test with install_flags option
	installFlagsData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypePacman,
		Opts: &map[string]any{
			"install_flags": "--asdeps",
		},
	}
	installerWithInstallFlags := newTestPacmanInstaller(installFlagsData)
	optsWithInstallFlags := installerWithInstallFlags.GetOpts()
	if optsWithInstallFlags.InstallFlags == nil || *optsWithInstallFlags.InstallFlags != "--asdeps" {
		t.Errorf("expected InstallFlags to be '--asdeps'")
	}

	// Test with update_flags option
	updateFlagsData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypePacman,
		Opts: &map[string]any{
			"update_flags": "--ignore vim",
		},
	}
	installerWithUpdateFlags := newTestPacmanInstaller(updateFlagsData)
	optsWithUpdateFlags := installerWithUpdateFlags.GetOpts()
	if optsWithUpdateFlags.UpdateFlags == nil || *optsWithUpdateFlags.UpdateFlags != "--ignore vim" {
		t.Errorf("expected UpdateFlags to be '--ignore vim'")
	}

	// Test with all flags options combined
	allFlagsData := &appconfig.InstallerData{
		Name: strPtr("vim"),
		Type: appconfig.InstallerTypePacman,
		Opts: &map[string]any{
			"flags":         "--common",
			"install_flags": "--install-specific",
			"update_flags":  "--update-specific",
		},
	}
	installerWithAllFlags := newTestPacmanInstaller(allFlagsData)
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
