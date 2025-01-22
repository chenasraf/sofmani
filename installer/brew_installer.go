package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type BrewInstaller struct {
	InstallerBase
	Config *appconfig.AppConfig
	Info   *appconfig.InstallerData
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
	return i.RunCmdPassThrough("brew", "install", name)
}

// Update implements IInstaller.
func (i *BrewInstaller) Update() error {
	return i.RunCmdPassThrough("brew", "upgrade", *i.Info.Name)
}

// CheckNeedsUpdate implements IInstaller.
func (i *BrewInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	success, err := i.RunCmdGetSuccess("brew", "outdated", "--json", *i.Info.Name)
	if err != nil {
		return false, err
	}
	return !success, nil
}

// CheckIsInstalled implements IInstaller.
func (i *BrewInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return i.RunCmdGetSuccess(utils.GetShellWhich(), i.GetBinName())
}

// GetData implements IInstaller.
func (i *BrewInstaller) GetData() *appconfig.InstallerData {
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
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

func NewBrewInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *BrewInstaller {
	i := &BrewInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Info:          installer,
	}

	return i
}
