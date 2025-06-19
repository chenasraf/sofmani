package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type AptInstaller struct {
	InstallerBase
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
		InstallerBase:  InstallerBase{Data: installer},
		Config:         cfg,
		Info:           installer,
		PackageManager: packageManager,
	}

	return i
}
