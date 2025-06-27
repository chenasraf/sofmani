package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

// NpmInstaller is an installer for npm, pnpm, and yarn packages.
type NpmInstaller struct {
	InstallerBase
	// Config is the application configuration.
	Config *appconfig.AppConfig
	// PackageManager is the package manager to use (npm, pnpm, or yarn).
	PackageManager NpmPackageManager
	// Info is the installer data.
	Info *appconfig.InstallerData
}

// NpmOpts represents options for the NpmInstaller.
type NpmOpts struct {
	//
}

// NpmPackageManager represents a Node.js package manager type.
// This type is also defined in apt_installer.go. Consider refactoring to a common location if appropriate.
type NpmPackageManager string

// Constants for supported Node.js package managers.
const (
	PackageManagerNpm  NpmPackageManager = "npm"  // PackageManagerNpm represents the npm package manager.
	PackageManagerYarn NpmPackageManager = "yarn" // PackageManagerYarn represents the yarn package manager.
	PackageManagerPnpm NpmPackageManager = "pnpm" // PackageManagerPnpm represents the pnpm package manager.
)

// Validate validates the installer configuration.
func (i *NpmInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	return errors
}

// Install implements IInstaller.
func (i *NpmInstaller) Install() error {
	return i.RunCmdPassThrough(string(i.PackageManager), "install", "--global", *i.Info.Name)
}

// Update implements IInstaller.
func (i *NpmInstaller) Update() error {
	return i.RunCmdPassThrough(string(i.PackageManager), "install", "--global", *i.Info.Name+"@latest")
}

// CheckNeedsUpdate implements IInstaller.
func (i *NpmInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	success, err := i.RunCmdGetSuccess(string(i.PackageManager), "outdated", "--global", "--json", *i.Info.Name)
	if err != nil {
		return false, err
	}
	return !success, nil
}

// CheckIsInstalled implements IInstaller.
func (i *NpmInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return i.RunCmdGetSuccess(utils.GetShellWhich(), i.GetBinName())
}

// GetData implements IInstaller.
func (i *NpmInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

// GetOpts returns the parsed options for the NpmInstaller.
func (i *NpmInstaller) GetOpts() *NpmOpts {
	opts := &NpmOpts{}
	// info := i.Info
	return opts
}

// GetBinName returns the binary name for the installer.
// It uses the BinName from the installer data if provided, otherwise it uses the installer name.
func (i *NpmInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

// NewNpmInstaller creates a new NpmInstaller.
func NewNpmInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *NpmInstaller {
	var packageManager NpmPackageManager
	switch installer.Type {
	case appconfig.InstallerTypeNpm:
		packageManager = PackageManagerNpm
	case appconfig.InstallerTypePnpm:
		packageManager = PackageManagerPnpm
	case appconfig.InstallerTypeYarn:
		packageManager = PackageManagerYarn
	}
	i := &NpmInstaller{
		InstallerBase:  InstallerBase{Data: installer},
		Config:         cfg,
		PackageManager: packageManager,
		Info:           installer,
	}

	return i
}
