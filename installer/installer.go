package installer

import (
	"fmt"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/appconfig/installertype"
)

type IInstaller interface {
	GetInfo() *appconfig.Installer
	CheckIsInstalled() (error, bool)
	CheckNeedsUpdate() (error, bool)
	Install() error
	Update() error
}

func GetInstaller(config *appconfig.AppConfig, installer *appconfig.Installer) (error, IInstaller) {
	switch installer.Type {
	case installertype.Brew:
		return nil, NewBrewInstaller(config, installer)
	}
	return fmt.Errorf("Installer type %s is not supported", installer.Type), nil
}

func RunInstaller(config *appconfig.AppConfig, installer IInstaller) error {
	err, installed := installer.CheckIsInstalled()
	if err != nil {
		return err
	}
	if installed {
		fmt.Printf("%s is already installed\n", installer.GetInfo().Name)
		// if *config.CheckUpdates {
		err, needsUpdate := installer.CheckNeedsUpdate()
		if err != nil {
			return err
		}
		if needsUpdate {
			fmt.Printf("%s needs an update\n", installer.GetInfo().Name)
			installer.Update()
		}
	} else {
		return nil
	}
	// }
	installer.Install()
	return nil
}
