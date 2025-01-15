package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type NpmInstaller struct {
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
	return utils.RunCmdPassThrough(i.Info.Environ(), string(i.PackageManager), "install", "--global", *i.Info.Name)
}

// Update implements IInstaller.
func (i *NpmInstaller) Update() error {
	return utils.RunCmdPassThrough(i.Info.Environ(), string(i.PackageManager), "install", "--global", *i.Info.Name+"@latest")
}

// CheckNeedsUpdate implements IInstaller.
func (i *NpmInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetData().CheckHasUpdate != nil {
		return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetShellWhich(), utils.GetOSShellArgs(*i.GetData().CheckHasUpdate)...)
	}
	err, success := utils.RunCmdGetSuccess(i.Info.Environ(), string(i.PackageManager), "outdated", "--global", "--json", *i.Info.Name)
	if err != nil {
		return err, false
	}
	return nil, !success
}

// CheckIsInstalled implements IInstaller.
func (i *NpmInstaller) CheckIsInstalled() (error, bool) {
	if i.GetData().CheckInstalled != nil {
		return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetOSShell(i.GetData().EnvShell), utils.GetOSShellArgs(*i.GetData().CheckInstalled)...)
	}
	return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetShellWhich(), i.GetBinName())
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
		Config:         cfg,
		PackageManager: packageManager,
		Info:           installer,
	}

	return i
}
