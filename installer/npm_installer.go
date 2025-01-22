package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type NpmInstaller struct {
	InstallerBase
	Config         *appconfig.AppConfig
	PackageManager PackageManager
	Info           *appconfig.InstallerData
}

type NpmOpts struct {
	//
}

type PackageManager string

const (
	PackageManagerNpm  PackageManager = "npm"
	PackageManagerYarn PackageManager = "yarn"
	PackageManagerPnpm PackageManager = "pnpm"
)

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

func (i *NpmInstaller) GetOpts() *NpmOpts {
	opts := &NpmOpts{}
	// info := i.Info
	return opts
}

func (i *NpmInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

func NewNpmInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *NpmInstaller {
	var packageManager PackageManager
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
