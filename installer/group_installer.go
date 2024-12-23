package installer

import (
	"os/exec"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
)

type GroupInstaller struct {
	Config *appconfig.AppConfig
	Info   *appconfig.Installer
}

type GroupOpts struct {
	BinName *string
}

// Install implements IInstaller.
func (i *GroupInstaller) Install() error {
	logger.Debug("Installing group %s", i.Info.Name)
	for _, step := range *i.Info.Steps {
		err, installer := GetInstaller(i.Config, &step)
		if err != nil {
			return err
		}
		RunInstaller(i.Config, installer)
	}
	return nil
}

// Update implements IInstaller.
func (i *GroupInstaller) Update() error {
	return nil
}

// CheckNeedsUpdate implements IInstaller.
func (i *GroupInstaller) CheckNeedsUpdate() (error, bool) {
	return nil, false
}

// CheckIsInstalled implements IInstaller.
func (i *GroupInstaller) CheckIsInstalled() (error, bool) {
	cmd := exec.Command("which", i.GetBinName())
	err := cmd.Run()
	if err != nil {
		return nil, false
	}
	return nil, true
}

// GetInfo implements IInstaller.
func (i *GroupInstaller) GetInfo() *appconfig.Installer {
	return i.Info
}

func (i *GroupInstaller) GetOpts() *GroupOpts {
	opts := &GroupOpts{}
	info := i.Info
	if info.Opts != nil {
		if binName, ok := (*info.Opts)["bin_name"].(string); ok {
			opts.BinName = &binName
		}
	}
	return opts
}

func (i *GroupInstaller) GetBinName() string {
	opts := i.GetOpts()
	if opts.BinName != nil && len(*opts.BinName) > 0 {
		return *opts.BinName
	}
	return i.Info.Name
}

func NewGroupInstaller(cfg *appconfig.AppConfig, installer *appconfig.Installer) *GroupInstaller {
	return &GroupInstaller{
		Config: cfg,
		Info:   installer,
	}
}
