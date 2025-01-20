package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type AptInstaller struct {
	Config         *appconfig.AppConfig
	Info           *appconfig.InstallerData
	PackageManager PackageManager
}

type AptOpts struct {
	//
}

const (
	PackageManagerApk PackageManager = "apk"
	PackageManagerApt PackageManager = "apt"
)

// Install implements IInstaller.
func (i *AptInstaller) Install() error {
	name := *i.Info.Name
	err := utils.RunCmdPassThrough(i.Info.Environ(), string(i.PackageManager), "update")
	if err != nil {
		return err
	}
	install := "install"
	if i.PackageManager == PackageManagerApk {
		install = "add"
	}
	return utils.RunCmdPassThrough(i.Info.Environ(), string(i.PackageManager), install, i.getConfirmArg(), name)
}

func (i *AptInstaller) getConfirmArg() string {
	confirm := "-y"
	if i.PackageManager == PackageManagerApk {
		confirm = ""
	}
	return confirm
}

// Update implements IInstaller.
func (i *AptInstaller) Update() error {
	return utils.RunCmdPassThrough(i.Info.Environ(), string(i.PackageManager), "upgrade", i.getConfirmArg(), *i.Info.Name)
}

// CheckNeedsUpdate implements IInstaller.
func (i *AptInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetData().CheckHasUpdate != nil {
		return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetOSShell(i.GetData().EnvShell), utils.GetOSShellArgs(*i.GetData().CheckHasUpdate)...)
	}
	err := utils.RunCmdPassThrough(i.Info.Environ(), "apk", "update")
	if err != nil {
		return err, false
	}
	err, success := utils.RunCmdGetSuccess(i.Info.Environ(), string(i.PackageManager), "--simulate", "upgrade", *i.Info.Name)
	if err != nil {
		return err, false
	}
	return nil, !success
}

// CheckIsInstalled implements IInstaller.
func (i *AptInstaller) CheckIsInstalled() (error, bool) {
	return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetShellWhich(), i.GetBinName())
}

// GetData implements IInstaller.
func (i *AptInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

func (i *AptInstaller) GetOpts() *AptOpts {
	opts := &AptOpts{}
	info := i.Info
	if info.Opts != nil {
		//
	}
	return opts
}

func (i *AptInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

func NewAptInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *AptInstaller {
	var packageManager PackageManager
	switch installer.Type {
	case appconfig.InstallerTypeApt:
		packageManager = PackageManagerApt
	case appconfig.InstallerTypeApk:
		packageManager = PackageManagerApk
	}
	i := &AptInstaller{
		Config:         cfg,
		Info:           installer,
		PackageManager: packageManager,
	}

	return i
}
