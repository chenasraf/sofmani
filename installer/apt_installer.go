package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type AptInstaller struct {
	Config *appconfig.AppConfig
	Info   *appconfig.InstallerData
}

type AptOpts struct {
	//
}

// Install implements IInstaller.
func (i *AptInstaller) Install() error {
	name := *i.Info.Name
	err := utils.RunCmdPassThrough(i.Info.Environ(), "apt", "update")
	if err != nil {
		return err
	}
	return utils.RunCmdPassThrough(i.Info.Environ(), "apt", "install", "-y", name)
}

// Update implements IInstaller.
func (i *AptInstaller) Update() error {
	return utils.RunCmdPassThrough(i.Info.Environ(), "apt", "upgrade", "-y", *i.Info.Name)
}

// CheckNeedsUpdate implements IInstaller.
func (i *AptInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetData().CheckHasUpdate != nil {
		return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetOSShell(i.GetData().EnvShell), utils.GetOSShellArgs(*i.GetData().CheckHasUpdate)...)
	}
	err, success := utils.RunCmdGetSuccess(i.Info.Environ(), "apt", "--simulate", "upgrade", *i.Info.Name)
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
	i := &AptInstaller{
		Config: cfg,
		Info:   installer,
	}

	return i
}
