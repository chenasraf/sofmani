package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/utils"
)

type GroupInstaller struct {
	Config *appconfig.AppConfig
	Info   *appconfig.Installer
}

type GroupOpts struct {
	//
}

// Install implements IInstaller.
func (i *GroupInstaller) Install() error {
	info := i.GetInfo()
	name := *info.Name
	logger.Debug("Installing group %s", name)
	for _, step := range *i.Info.Steps {
		err, installer := GetInstaller(i.Config, &step)
		if err != nil {
			return err
		}
		if installer == nil {
			logger.Warn("Installer type %s is not supported, skipping", step.Type)
		} else {
			RunInstaller(i.Config, installer)
		}
	}
	return nil
}

// Update implements IInstaller.
func (i *GroupInstaller) Update() error {
	return nil
}

// CheckNeedsUpdate implements IInstaller.
func (i *GroupInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetInfo().CheckHasUpdate != nil {
		return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetOSShell(i.GetInfo().EnvShell), utils.GetOSShellArgs(*i.GetInfo().CheckHasUpdate)...)
	}
	return nil, true
}

// CheckIsInstalled implements IInstaller.
func (i *GroupInstaller) CheckIsInstalled() (error, bool) {
	if i.GetInfo().CheckInstalled != nil {
		return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetOSShell(i.GetInfo().EnvShell), utils.GetOSShellArgs(*i.GetInfo().CheckInstalled)...)
	}
	return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetShellWhich(), i.GetBinName())
}

// GetInfo implements IInstaller.
func (i *GroupInstaller) GetInfo() *appconfig.Installer {
	return i.Info
}

func (i *GroupInstaller) GetOpts() *GroupOpts {
	opts := &GroupOpts{}
	info := i.Info
	if info.Opts != nil {
		//
	}
	return opts
}

func (i *GroupInstaller) GetBinName() string {
	info := i.GetInfo()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

func NewGroupInstaller(cfg *appconfig.AppConfig, installer *appconfig.Installer) *GroupInstaller {
	return &GroupInstaller{
		Config: cfg,
		Info:   installer,
	}
}
