package installer

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/chenasraf/sofmani/appconfig"
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

	cmd := exec.Command("sh", "-c", tmpfile)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	cmd.Wait()
	return nil
}

// Update implements IInstaller.
func (i *ShellInstaller) Update() error {
	return i.Install()
}

// CheckNeedsUpdate implements IInstaller.
func (i *ShellInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetOpts().CheckHasUpdate != nil {
		cmd := exec.Command("sh", "-c", *i.GetOpts().CheckHasUpdate)
		err := cmd.Run()
		if err != nil {
			return err, true
		}
		return nil, false
	}
	return nil, false
}

// CheckIsInstalled implements IInstaller.
func (i *ShellInstaller) CheckIsInstalled() (error, bool) {
	// cmd := exec.Command("brew", "list", i.Info.Name)
	cmd := exec.Command("which", i.GetBinName())
	err := cmd.Run()
	if err != nil {
		return nil, false
	}
	return nil, true
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
