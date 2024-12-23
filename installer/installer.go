package installer

import (
	"fmt"
	"runtime"

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
	case appconfig.InstallerTypeShell:
		return nil, NewShellInstaller(config, installer)
	}
	return fmt.Errorf("Installer type %s is not supported", installer.Type), nil
}

func GetCurrentPlatform() appconfig.Platform {
	switch runtime.GOOS {
	case "darwin":
		return appconfig.PlatformMacos
	case "linux":
		return appconfig.PlatformLinux
	case "windows":
		return appconfig.PlatformWindows
	}
	panic(fmt.Sprintf("Unsupported platform %s", runtime.GOOS))
}

func RunInstaller(config *appconfig.AppConfig, installer IInstaller) error {
	logger.Debug("Checking if %s should run on %s", installer.GetInfo().Name, GetCurrentPlatform())
	curOS := GetCurrentPlatform()
	if !GetShouldRunOnOS(installer, curOS) {
		logger.Debug("%s should not run on %s, skipping", installer.GetInfo().Name, curOS)
		return nil
	}
	logger.Debug("Checking if %s is installed", installer.GetInfo().Name)
	err, installed := installer.CheckIsInstalled()
	if err != nil {
		return err
	}
	if installed {
		logger.Debug("%s is already installed", installer.GetInfo().Name)
		if config.CheckUpdates {
			logger.Debug("Checking if %s needs an update", installer.GetInfo().Name)
			err, needsUpdate := installer.CheckNeedsUpdate()
			if err != nil {
				return err
			}
			if needsUpdate {
				logger.Info("%s has an update", installer.GetInfo().Name)
				installer.Update()
			}
		} else {
			return nil
		}
	}
	logger.Info("Installing %s (%s)", installer.GetInfo().Name, installer.GetInfo().Type)
	err = installer.Install()
	if err != nil {
		return err
	}
	return nil
}

func GetShouldRunOnOS(installer IInstaller, curOS appconfig.Platform) bool {
	platforms := installer.GetInfo().Platforms
	if platforms == nil {
		return true
	}

	if platforms.Only != nil {
		logger.Debug("Checking if %s is in %s", curOS, platforms.Only)
		return appconfig.ContainsPlatform(platforms.Only, curOS)
	}
	if platforms.Except != nil {
		logger.Debug("Checking if %s is not in %s", curOS, platforms.Except)
		return !appconfig.ContainsPlatform(platforms.Except, curOS)
	}
	return true
}
