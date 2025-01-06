package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type BrewInstaller struct {
	Config *appconfig.AppConfig
	Info   *appconfig.Installer
}

type BrewOpts struct {
	Tap *string
}

// Install implements IInstaller.
func (i *BrewInstaller) Install() error {
	name := *i.Info.Name
	if i.GetOpts().Tap != nil {
		name = *i.GetOpts().Tap + "/" + name
	}
	return utils.RunCmdPassThrough(i.Info.Environ(), "brew", "install", name)
}

// Update implements IInstaller.
func (i *BrewInstaller) Update() error {
	return utils.RunCmdPassThrough(i.Info.Environ(), "brew", "upgrade", *i.Info.Name)
}

// CheckNeedsUpdate implements IInstaller.
func (i *BrewInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetInfo().CheckHasUpdate != nil {
		return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetOSShell(i.GetInfo().EnvShell), utils.GetOSShellArgs(*i.GetInfo().CheckHasUpdate)...)
	}
	err, success := utils.RunCmdGetSuccess(i.Info.Environ(), "brew", "outdated", "--json", *i.Info.Name)
	if err != nil {
		return err, false
	}
	return nil, !success
}

// CheckIsInstalled implements IInstaller.
func (i *BrewInstaller) CheckIsInstalled() (error, bool) {
	return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetShellWhich(), i.GetBinName())
}

// GetInfo implements IInstaller.
func (i *BrewInstaller) GetInfo() *appconfig.Installer {
	return i.Info
}

func (i *BrewInstaller) GetOpts() *BrewOpts {
	opts := &BrewOpts{}
	info := i.Info
	if info.Opts != nil {
		if tap, ok := (*info.Opts)["tap"].(string); ok {
			opts.Tap = &tap
		}
	}
	return opts
}

func (i *BrewInstaller) GetBinName() string {
	info := i.GetInfo()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

func NewBrewInstaller(cfg *appconfig.AppConfig, installer *appconfig.Installer) *BrewInstaller {
	i := &BrewInstaller{
		Config: cfg,
		Info:   installer,
	}

	return i
}
