package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type ShellInstaller struct {
	Config *appconfig.AppConfig
	Info   *appconfig.InstallerData
}

type ShellOpts struct {
	Command       *string
	UpdateCommand *string
}

// Install implements IInstaller.
func (i *ShellInstaller) Install() error {
	return utils.RunCmdAsFile(i.Info.Environ(), *i.GetOpts().Command, i.GetData().EnvShell)
}

// Update implements IInstaller.
func (i *ShellInstaller) Update() error {
	if i.GetOpts().UpdateCommand != nil {
		return utils.RunCmdAsFile(i.Info.Environ(), *i.GetOpts().UpdateCommand, i.GetData().EnvShell)
	}
	return i.Install()
}

// CheckNeedsUpdate implements IInstaller.
func (i *ShellInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetData().CheckHasUpdate != nil {
		shell := utils.GetOSShell(i.GetData().EnvShell)
		args := utils.GetOSShellArgs(*i.GetData().CheckHasUpdate)
		return utils.RunCmdGetSuccess(i.Info.Environ(), shell, args...)
	}
	return nil, false
}

// CheckIsInstalled implements IInstaller.
func (i *ShellInstaller) CheckIsInstalled() (error, bool) {
	if i.GetData().CheckInstalled != nil {
		shell := utils.GetOSShell(i.GetData().EnvShell)
		args := utils.GetOSShellArgs(*i.GetData().CheckInstalled)
		return utils.RunCmdGetSuccess(i.Info.Environ(), shell, args...)
	}
	return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetShellWhich(), i.GetBinName())
}

// GetData implements IInstaller.
func (i *ShellInstaller) GetData() *appconfig.InstallerData {
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
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

func NewShellInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *ShellInstaller {
	return &ShellInstaller{
		Config: cfg,
		Info:   installer,
	}
}
