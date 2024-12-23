package installer

import (
	"fmt"
	"os"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type ShellInstaller struct {
	Config *appconfig.AppConfig
	Info   *appconfig.Installer
}

type ShellOpts struct {
	Command *string
}

// Install implements IInstaller.
func (i *ShellInstaller) Install() error {
	tmpdir := os.TempDir()
	tmpfile := fmt.Sprintf("%s%s", tmpdir, "install.sh")
	commandStr := fmt.Sprintf("#!/bin/bash\n%s\n", *i.GetOpts().Command)
	err := os.WriteFile(tmpfile, []byte(commandStr), 0755)
	if err != nil {
		return err
	}

	return utils.RunCmdPassThrough("sh", "-c", tmpfile)
}

// Update implements IInstaller.
func (i *ShellInstaller) Update() error {
	return i.Install()
}

// CheckNeedsUpdate implements IInstaller.
func (i *ShellInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetInfo().CheckHasUpdate != nil {
		return utils.RunCmdGetSuccess("sh", "-c", *i.GetInfo().CheckHasUpdate)
	}
	return nil, false
}

// CheckIsInstalled implements IInstaller.
func (i *ShellInstaller) CheckIsInstalled() (error, bool) {
	if i.GetInfo().CheckInstalled != nil {
		return utils.RunCmdGetSuccess("sh", "-c", *i.GetInfo().CheckInstalled)
	}
	return utils.RunCmdGetSuccess("which", i.GetBinName())
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
