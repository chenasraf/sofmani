package installer

import (
	"fmt"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
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
	case appconfig.InstallerTypeBrew:
		return nil, NewBrewInstaller(config, installer)
	case appconfig.InstallerTypeGroup:
		return nil, NewGroupInstaller(config, installer)
	}
	return fmt.Errorf("Installer type %s is not supported", installer.Type), nil
}

func RunInstaller(config *appconfig.AppConfig, installer IInstaller) error {
	logger.Info("Checking if %s is installed", installer.GetInfo().Name)
	err, installed := installer.CheckIsInstalled()
	if err != nil {
		return err
	}
	if installed {
		logger.Info("%s is already installed", installer.GetInfo().Name)
		if config.CheckUpdates {
			logger.Info("Checking if %s needs an update", installer.GetInfo().Name)
			err, needsUpdate := installer.CheckNeedsUpdate()
			if err != nil {
				return err
			}
			if needsUpdate {
				logger.Info("%s needs an update", installer.GetInfo().Name)
				installer.Update()
			}
		} else {
			return nil
		}
	}
	err = installer.Install()
	if err != nil {
		return err
	}
	return nil
}
