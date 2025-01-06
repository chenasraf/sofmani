package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type ShellInstaller struct {
	Config *appconfig.AppConfig
	Info   *appconfig.Installer
}

type ShellOpts struct {
	Command       *string
	UpdateCommand *string
}

// Install implements IInstaller.
func (i *ShellInstaller) Install() error {
	return utils.RunCmdAsFile(i.Info.Environ(), *i.GetOpts().Command, i.GetInfo().EnvShell)
}

// Update implements IInstaller.
func (i *ShellInstaller) Update() error {
	if i.GetOpts().UpdateCommand != nil {
		return utils.RunCmdAsFile(i.Info.Environ(), *i.GetOpts().UpdateCommand, i.GetInfo().EnvShell)
	}
	return i.Install()
}

// CheckNeedsUpdate implements IInstaller.
func (i *ShellInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetInfo().CheckHasUpdate != nil {
		shell := utils.GetOSShell(i.GetInfo().EnvShell)
		args := utils.GetOSShellArgs(*i.GetInfo().CheckHasUpdate)
		return utils.RunCmdGetSuccess(i.Info.Environ(), shell, args...)
	}
	return nil, false
}

// CheckIsInstalled implements IInstaller.
func (i *ShellInstaller) CheckIsInstalled() (error, bool) {
	if i.GetInfo().CheckInstalled != nil {
		shell := utils.GetOSShell(i.GetInfo().EnvShell)
		args := utils.GetOSShellArgs(*i.GetInfo().CheckInstalled)
		return utils.RunCmdGetSuccess(i.Info.Environ(), shell, args...)
	}
	return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetShellWhich(), i.GetBinName())
}

// GetInfo implements IInstaller.
func (i *ShellInstaller) GetInfo() *appconfig.Installer {
	return i.Info
}

func (i *ShellInstaller) GetOpts() *ShellOpts {
	opts := &ShellOpts{}
	info := i.Info
	if info.Opts != nil {
		if command, ok := (*info.Opts)["command"].(string); ok {
			opts.Command = &command
		}
		if updateCommand, ok := (*info.Opts)["update_command"].(string); ok {
			opts.UpdateCommand = &updateCommand
		}
	}
	return opts
}

func (i *ShellInstaller) GetBinName() string {
	info := i.GetInfo()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

func NewShellInstaller(cfg *appconfig.AppConfig, installer *appconfig.Installer) *ShellInstaller {
	return &ShellInstaller{
		Config: cfg,
		Info:   installer,
	}
}
