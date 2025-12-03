package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
)

// PacmanInstaller is an installer for pacman and yay packages.
type PacmanInstaller struct {
	InstallerBase
	// Config is the application configuration.
	Config *appconfig.AppConfig
	// Info is the installer data.
	Info *appconfig.InstallerData
	// PackageManager is the package manager to use (pacman or yay).
	PackageManager PacmanPackageManager
}

// PacmanOpts represents options for the PacmanInstaller.
type PacmanOpts struct {
	// Needed skips reinstalling up-to-date packages (--needed flag).
	Needed *bool
}

// PacmanPackageManager represents an Arch Linux package manager type.
type PacmanPackageManager string

// Constants for supported Arch Linux package managers.
const (
	PackageManagerPacman PacmanPackageManager = "pacman" // PackageManagerPacman represents the pacman package manager.
	PackageManagerYay    PacmanPackageManager = "yay"    // PackageManagerYay represents the yay AUR helper.
)

// Validate validates the installer configuration.
func (i *PacmanInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	return errors
}

// Install implements IInstaller.
func (i *PacmanInstaller) Install() error {
	name := *i.Info.Name
	args := []string{"-S", "--noconfirm"}
	if i.GetOpts().Needed != nil && *i.GetOpts().Needed {
		args = append(args, "--needed")
	}
	args = append(args, name)
	return i.RunCmdPassThrough(string(i.PackageManager), args...)
}

// Update implements IInstaller.
func (i *PacmanInstaller) Update() error {
	name := *i.Info.Name
	args := []string{"-S", "--noconfirm"}
	if i.GetOpts().Needed != nil && *i.GetOpts().Needed {
		args = append(args, "--needed")
	}
	args = append(args, name)
	return i.RunCmdPassThrough(string(i.PackageManager), args...)
}

// CheckNeedsUpdate implements IInstaller.
func (i *PacmanInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	// -Qu lists packages that have updates available
	// If the package has an update, it will be in the output
	output, err := i.RunCmdGetOutput(string(i.PackageManager), "-Qu", *i.Info.Name)
	if err != nil {
		// No output or error means no updates needed
		return false, nil
	}
	// If we got output, there are updates available
	return len(output) > 0, nil
}

// CheckIsInstalled implements IInstaller.
func (i *PacmanInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	// Use pacman -Q to check if package is installed (works for all packages including fonts/libraries)
	return i.RunCmdGetSuccess(string(i.PackageManager), "-Q", *i.Info.Name)
}

// GetData implements IInstaller.
func (i *PacmanInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

// GetOpts returns the parsed options for the PacmanInstaller.
func (i *PacmanInstaller) GetOpts() *PacmanOpts {
	opts := &PacmanOpts{}
	info := i.Info
	if info.Opts != nil {
		if needed, ok := (*info.Opts)["needed"].(bool); ok {
			opts.Needed = &needed
		}
	}
	return opts
}

// GetBinName returns the binary name for the installer.
// It uses the BinName from the installer data if provided, otherwise it uses the installer name.
func (i *PacmanInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

// NewPacmanInstaller creates a new PacmanInstaller.
func NewPacmanInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *PacmanInstaller {
	var packageManager PacmanPackageManager
	switch installer.Type {
	case appconfig.InstallerTypePacman:
		packageManager = PackageManagerPacman
	case appconfig.InstallerTypeYay:
		packageManager = PackageManagerYay
	}
	i := &PacmanInstaller{
		InstallerBase:  InstallerBase{Data: installer},
		Config:         cfg,
		Info:           installer,
		PackageManager: packageManager,
	}

	return i
}
