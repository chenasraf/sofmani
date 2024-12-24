package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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
	tmpfile := i.getShellScript(tmpdir)
	commandStr, err := i.getScriptContents(*i.GetOpts().Command)
	if err != nil {
		return err
	}
	err = os.WriteFile(tmpfile, []byte(commandStr), 0755)
	if err != nil {
		return err
	}

	shell := utils.GetOSShell()
	args := utils.GetOSShellArgs(tmpfile)
	return utils.RunCmdPassThrough(i.Info.Environ(), shell, args...)
}

// Update implements IInstaller.
func (i *ShellInstaller) Update() error {
	return i.Install()
}

// CheckNeedsUpdate implements IInstaller.
func (i *ShellInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetInfo().CheckHasUpdate != nil {
		shell := utils.GetOSShell()
		args := utils.GetOSShellArgs(*i.GetInfo().CheckHasUpdate)
		return utils.RunCmdGetSuccess(i.Info.Environ(), shell, args...)
	}
	return nil, false
}

// CheckIsInstalled implements IInstaller.
func (i *ShellInstaller) CheckIsInstalled() (error, bool) {
	if i.GetInfo().CheckInstalled != nil {
		shell := utils.GetOSShell()
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

func (*ShellInstaller) getShellScript(dir string) string {
	var filename string
	switch runtime.GOOS {
	case "windows":
		filename = "install.bat"
	case "linux", "darwin":
		filename = "install.sh"
	}
	tmpfile := filepath.Join(dir, filename)
	return tmpfile
}

func (i *ShellInstaller) getScriptContents(script string) (string, error) {
	switch runtime.GOOS {
	case "windows":
		return *i.GetOpts().Command, nil
	case "linux", "darwin":
		return fmt.Sprintf("#!/usr/bin/i.Info.Environ() bash\n%s\n", script), nil
	}
	return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func NewShellInstaller(cfg *appconfig.AppConfig, installer *appconfig.Installer) *ShellInstaller {
	return &ShellInstaller{
		Config: cfg,
		Info:   installer,
	}
}
