package installer

import (
	"fmt"

	"github.com/chenasraf/sofmani/appconfig"
)

type GroupInstaller struct {
	Config *appconfig.AppConfig
	Info   *appconfig.Installer
}

// Install implements IInstaller.
func (i *GroupInstaller) Install() error {
	fmt.Printf("Installing group %s\n", i.Info.Name)
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
	return nil, false
}

// GetInfo implements IInstaller.
func (i *GroupInstaller) GetInfo() *appconfig.Installer {
	return i.Info
}

func NewGroupInstaller(cfg *appconfig.AppConfig, installer *appconfig.Installer) *GroupInstaller {
	return &GroupInstaller{
		Config: cfg,
		Info:   installer,
	}
}
