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
	Command        *string
	BinName        *string
	CheckHasUpdate *string
}

// Install implements IInstaller.
func (i *ShellInstaller) Install() error {
	tmpdir := os.TempDir()
	tmpfile := fmt.Sprintf("%s/%s", tmpdir, "install.sh")
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
	if i.GetOpts().CheckHasUpdate != nil {
		return utils.RunCmdGetSuccess("sh", "-c", *i.GetOpts().CheckHasUpdate)
	}
	return nil, false
}

// CheckIsInstalled implements IInstaller.
func (i *ShellInstaller) CheckIsInstalled() (error, bool) {
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
		if binName, ok := (*info.Opts)["bin_name"].(string); ok {
			opts.BinName = &binName
		}
		if command, ok := (*info.Opts)["check_has_update"].(string); ok {
			opts.CheckHasUpdate = &command
		}
	}
	return opts
}

func (i *ShellInstaller) GetBinName() string {
	opts := i.GetOpts()
	if opts.BinName != nil && len(*opts.BinName) > 0 {
		return *opts.BinName
	}
	return i.Info.Name
}

func NewShellInstaller(cfg *appconfig.AppConfig, installer *appconfig.Installer) *ShellInstaller {
	return &ShellInstaller{
		Config: cfg,
		Info:   installer,
	}
}
