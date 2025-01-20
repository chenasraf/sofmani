package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type PipxInstaller struct {
	Config *appconfig.AppConfig
	Info   *appconfig.InstallerData
}

type PipxOpts struct {
	//
}

// Install implements IInstaller.
func (i *PipxInstaller) Install() error {
	name := *i.Info.Name
	return utils.RunCmdPassThrough(i.Info.Environ(), "pipx", "install", name)
}

// Update implements IInstaller.
func (i *PipxInstaller) Update() error {
	return utils.RunCmdPassThrough(i.Info.Environ(), "pipx", "upgrade", *i.Info.Name)
}

// CheckNeedsUpdate implements IInstaller.
func (i *PipxInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetData().CheckHasUpdate != nil {
		return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetOSShell(i.GetData().EnvShell), utils.GetOSShellArgs(*i.GetData().CheckHasUpdate)...)
	}
	err, success := utils.RunCmdGetSuccess(i.Info.Environ(), "pipx", "upgrade", "--pip-args=--dry-run", *i.Info.Name)
	if err != nil {
		return err, false
	}
	return nil, !success
}

// CheckIsInstalled implements IInstaller.
func (i *PipxInstaller) CheckIsInstalled() (error, bool) {
	return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetShellWhich(), i.GetBinName())
}

// GetData implements IInstaller.
func (i *PipxInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

func (i *PipxInstaller) GetOpts() *PipxOpts {
	opts := &PipxOpts{}
	info := i.Info
	if info.Opts != nil {
		//
	}
	return opts
}

func (i *PipxInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

func NewPipxInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *PipxInstaller {
	i := &PipxInstaller{
		Config: cfg,
		Info:   installer,
	}

	return i
}
