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
}
