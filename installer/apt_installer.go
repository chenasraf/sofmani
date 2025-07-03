package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

// AptInstaller is an installer for apt and apk packages.
type AptInstaller struct {
	InstallerBase
	// Config is the application configuration.
	Config *appconfig.AppConfig
	// Info is the installer data.
	Info *appconfig.InstallerData
	// PackageManager is the package manager to use (apt or apk).
	PackageManager AptPackageManager
}

// AptOpts represents options for the AptInstaller.
type AptOpts struct {
	//
}

// AptPackageManager represents a package manager type.
type AptPackageManager string

// Constants for supported package managers.
const (
	PackageManagerApk AptPackageManager = "apk" // PackageManagerApk represents the apk package manager.
	PackageManagerApt AptPackageManager = "apt" // PackageManagerApt represents the apt package manager.
)

// Validate validates the installer configuration.
func (i *AptInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	return errors
}

// Install implements IInstaller.
func (i *AptInstaller) Install() error {
	name := *i.Info.Name
	err := i.RunCmdPassThrough(string(i.PackageManager), "update")
	if err != nil {
		return err
	}
	install := "install"
	if i.PackageManager == PackageManagerApk {
		install = "add"
	}
	return i.RunCmdPassThrough(string(i.PackageManager), install, i.getConfirmArg(), name)
}

// getConfirmArg returns the appropriate confirmation argument for the package manager.
// For apt, it returns "-y". For apk, it returns an empty string.
func (i *AptInstaller) getConfirmArg() string {
	confirm := "-y"
	if i.PackageManager == PackageManagerApk {
		confirm = ""
	}
	return confirm
}

// Update implements IInstaller.
func (i *AptInstaller) Update() error {
	return i.RunCmdPassThrough(string(i.PackageManager), "upgrade", i.getConfirmArg(), *i.Info.Name)
}

// CheckNeedsUpdate implements IInstaller.
func (i *AptInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	err := i.RunCmdPassThrough("apk", "update")
	if err != nil {
		return false, err
	}
	success, err := i.RunCmdGetSuccess(string(i.PackageManager), "--simulate", "upgrade", *i.Info.Name)
	if err != nil {
		return false, err
	}
	return !success, nil
}

// CheckIsInstalled implements IInstaller.
func (i *AptInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return i.RunCmdGetSuccess(utils.GetShellWhich(), i.GetBinName())
}

// GetData implements IInstaller.
func (i *AptInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

// GetOpts returns the parsed options for the AptInstaller.
func (i *AptInstaller) GetOpts() *AptOpts {
	return &AptOpts{}
}

// GetBinName returns the binary name for the installer.
// It uses the BinName from the installer data if provided, otherwise it uses the installer name.
func (i *AptInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

// NewAptInstaller creates a new AptInstaller.
func NewAptInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *AptInstaller {
	var packageManager AptPackageManager
	switch installer.Type {
	case appconfig.InstallerTypeApt:
		packageManager = PackageManagerApt
	case appconfig.InstallerTypeApk:
		packageManager = PackageManagerApk
	}
	i := &AptInstaller{
		InstallerBase:  InstallerBase{Data: installer},
		Config:         cfg,
		Info:           installer,
		PackageManager: packageManager,
	}

	return i
}
