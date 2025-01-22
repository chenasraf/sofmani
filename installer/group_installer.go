package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/utils"
)

type GroupInstaller struct {
	InstallerBase
	Config *appconfig.AppConfig
	Data   *appconfig.InstallerData
}

type GroupOpts struct {
	//
}

// Install implements IInstaller.
func (i *GroupInstaller) Install() error {
	info := i.GetData()
	name := *info.Name
	logger.Debug("Installing group %s", name)
	for _, step := range *i.Data.Steps {
		installer, err := GetInstaller(i.Config, &step)
		if err != nil {
			return err
		}
		if installer == nil {
			logger.Warn("Installer type %s is not supported, skipping", step.Type)
		} else {
			RunInstaller(i.Config, installer)
		}
	}
	return nil
}

// Update implements IInstaller.
func (i *GroupInstaller) Update() error {
	return i.Install()
}

// CheckNeedsUpdate implements IInstaller.
func (i *GroupInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	return true, nil
}

// CheckIsInstalled implements IInstaller.
func (i *GroupInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return i.RunCmdGetSuccess(utils.GetShellWhich(), i.GetBinName())
}

// GetData implements IInstaller.
func (i *GroupInstaller) GetData() *appconfig.InstallerData {
	return i.Data
}

func (i *GroupInstaller) GetOpts() *GroupOpts {
	opts := &GroupOpts{}
	info := i.GetData()
	if info.Opts != nil {
		//
	}
	return opts
}

func (i *GroupInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

func NewGroupInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *GroupInstaller {
	return &GroupInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Data:          installer,
	}
}
