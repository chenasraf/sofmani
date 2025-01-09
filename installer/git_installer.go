package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type GitInstaller struct {
	Config *appconfig.AppConfig
	Info   *appconfig.Installer
}

type GitOpts struct {
	Destination *string
	Ref         *string
}

// Install implements IInstaller.
func (i *GitInstaller) Install() error {
	args := []string{"clone", i.GetRepositoryUrl(), i.GetInstallDir()}
	err := utils.RunCmdPassThrough(i.Info.Environ(), "git", args...)
	if err != nil {
		return err
	}
	if i.GetOpts().Ref != nil {
		return utils.RunCmdPassThrough(i.Info.Environ(), "git", "-C", i.GetInstallDir(), "checkout", *i.GetOpts().Ref)
	}
	return nil
}

// Update implements IInstaller.
func (i *GitInstaller) Update() error {
	return utils.RunCmdPassThrough(i.Info.Environ(), "git", "-C", i.GetInstallDir(), "pull")
}

// CheckNeedsUpdate implements IInstaller.
func (i *GitInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetInfo().CheckHasUpdate != nil {
		return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetOSShell(i.GetInfo().EnvShell), utils.GetOSShellArgs(*i.GetInfo().CheckHasUpdate)...)
	}
	err, _ := utils.RunCmdGetSuccess(i.Info.Environ(), "git", "-C", i.GetInstallDir(), "fetch")
	if err != nil {
		return err, false
	}
	output, err := utils.RunCmdGetOutput(i.Info.Environ(), "git", "-C", i.GetInstallDir(), "status", "-uno")
	if err != nil {
		return err, false
	}
	if strings.Contains(string(output), "Your branch is behind") {
		return nil, true
	}
	return nil, false
}

// CheckIsInstalled implements IInstaller.
func (i *GitInstaller) CheckIsInstalled() (error, bool) {
	return utils.PathExists(i.GetInstallDir())
}

// GetInfo implements IInstaller.
func (i *GitInstaller) GetInfo() *appconfig.Installer {
	return i.Info
}

func (i *GitInstaller) GetOpts() *GitOpts {
	opts := &GitOpts{}
	info := i.Info
	if info.Opts != nil {
		if destination, ok := (*info.Opts)["destination"].(string); ok {
			destination = utils.GetRealPath(i.GetInfo().Environ(), destination)
			opts.Destination = &destination
		}
		if ref, ok := (*info.Opts)["ref"].(string); ok {
			opts.Ref = &ref
		}
	}
	return opts
}

func (i *GitInstaller) GetRepositoryUrl() string {
	info := i.Info
	name := *info.Name
	if strings.HasPrefix(name, "https://") || strings.HasPrefix(name, "git@") {
		return name
	}
	return fmt.Sprintf("https://github.com/%s", name)
}

func (i *GitInstaller) GetDestination() string {
	if i.GetOpts().Destination != nil {
		return *i.GetOpts().Destination
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}

func (i *GitInstaller) GetInstallDir() string {
	return filepath.Join(i.GetDestination(), filepath.Base(*i.Info.Name))
}

func NewGitInstaller(cfg *appconfig.AppConfig, installer *appconfig.Installer) *GitInstaller {
	i := &GitInstaller{
		Config: cfg,
		Info:   installer,
	}

	return i
}
