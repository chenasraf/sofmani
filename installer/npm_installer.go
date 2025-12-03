package installer

import (
	"strings"

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
	// Flags is a string of additional flags to pass to the npm/pnpm/yarn command.
	Flags *string
	// InstallFlags is a string of additional flags to pass only during install.
	InstallFlags *string
	// UpdateFlags is a string of additional flags to pass only during update.
	UpdateFlags *string
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
	opts := i.GetOpts()
	args := []string{"install", "--global"}
	if opts.InstallFlags != nil {
		args = append(args, strings.Fields(*opts.InstallFlags)...)
	} else if opts.Flags != nil {
		args = append(args, strings.Fields(*opts.Flags)...)
	}
	args = append(args, *i.Info.Name)
	return i.RunCmdPassThrough(string(i.PackageManager), args...)
}

// Update implements IInstaller.
func (i *NpmInstaller) Update() error {
	opts := i.GetOpts()
	args := []string{"install", "--global"}
	if opts.UpdateFlags != nil {
		args = append(args, strings.Fields(*opts.UpdateFlags)...)
	} else if opts.Flags != nil {
		args = append(args, strings.Fields(*opts.Flags)...)
	}
	args = append(args, *i.Info.Name+"@latest")
	return i.RunCmdPassThrough(string(i.PackageManager), args...)
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
	info := i.Info
	if info.Opts != nil {
		if flags, ok := (*info.Opts)["flags"].(string); ok {
			opts.Flags = &flags
		}
		if installFlags, ok := (*info.Opts)["install_flags"].(string); ok {
			opts.InstallFlags = &installFlags
		}
		if updateFlags, ok := (*info.Opts)["update_flags"].(string); ok {
			opts.UpdateFlags = &updateFlags
		}
	}
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
