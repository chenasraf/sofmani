package installer

import (
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/utils"
)

type RsyncInstaller struct {
	Config *appconfig.AppConfig
	Info   *appconfig.Installer
}

type RsyncOpts struct {
	Source      *string
	Destination *string
	Flags       *string
}

// Install implements IInstaller.
func (i *RsyncInstaller) Install() error {
	defaultFlags := "-tr"
	if i.Config.Debug {
		defaultFlags += "v"
	}
	flags := []string{defaultFlags}
	if i.GetOpts().Flags != nil {
		for _, flag := range strings.Split(*i.GetOpts().Flags, " ") {
			flags = append(flags, flag)
		}
	}

	src := utils.GetRealPath(i.Info.Environ(), *i.GetOpts().Source)
	dest := utils.GetRealPath(i.Info.Environ(), *i.GetOpts().Destination)

	flags = append(flags, src)
	flags = append(flags, dest)

	logger.Debug("rsync %s to %s", src, dest)
	return utils.RunCmdPassThrough(i.Info.Environ(), "rsync", flags...)
}

// Update implements IInstaller.
func (i *RsyncInstaller) Update() error {
	return i.Install()
}

// CheckNeedsUpdate implements IInstaller.
func (i *RsyncInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetInfo().CheckHasUpdate != nil {
		return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetOSShell(i.GetInfo().EnvShell), utils.GetOSShellArgs(*i.GetInfo().CheckHasUpdate)...)
	}
	return nil, true
}

// CheckIsInstalled implements IInstaller.
func (i *RsyncInstaller) CheckIsInstalled() (error, bool) {
	if i.GetInfo().CheckInstalled != nil {
		return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetOSShell(i.GetInfo().EnvShell), utils.GetOSShellArgs(*i.GetInfo().CheckInstalled)...)
	}
	return nil, false
}

// GetInfo implements IInstaller.
func (i *RsyncInstaller) GetInfo() *appconfig.Installer {
	return i.Info
}

func (i *RsyncInstaller) GetOpts() *RsyncOpts {
	opts := &RsyncOpts{}
	info := i.Info
	if info.Opts != nil {
		if src, ok := (*info.Opts)["source"].(string); ok {
			opts.Source = &src
		}
		if dest, ok := (*info.Opts)["destination"].(string); ok {
			opts.Destination = &dest
		}
		if flags, ok := (*info.Opts)["flags"].(string); ok {
			opts.Flags = &flags
		}
	}
	return opts
}

func (i *RsyncInstaller) GetBinName() string {
	info := i.GetInfo()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

func NewRsyncInstaller(cfg *appconfig.AppConfig, installer *appconfig.Installer) *RsyncInstaller {
	i := &RsyncInstaller{
		Config: cfg,
		Info:   installer,
	}

	return i
}
