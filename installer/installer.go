package installer

import (
	"fmt"
	"runtime"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/utils"
)

type IInstaller interface {
	GetInfo() *appconfig.Installer
	CheckIsInstalled() (error, bool)
	CheckNeedsUpdate() (error, bool)
	Install() error
	Update() error
}

func GetInstaller(config *appconfig.AppConfig, installer *appconfig.Installer) (error, IInstaller) {
	installer = InstallerWithDefaults(installer, installer.Type, config.Defaults)
	switch installer.Type {
	case appconfig.InstallerTypeGroup:
		return nil, NewGroupInstaller(config, installer)
	case appconfig.InstallerTypeBrew:
		return nil, NewBrewInstaller(config, installer)
	case appconfig.InstallerTypeShell:
		return nil, NewShellInstaller(config, installer)
	case appconfig.InstallerTypeRsync:
		return nil, NewRsyncInstaller(config, installer)
	case appconfig.InstallerTypeNpm, appconfig.InstallerTypePnpm, appconfig.InstallerTypeYarn:
		return nil, NewNpmInstaller(config, installer)
	}
	return nil, nil
}

func InstallerWithDefaults(
	installer *appconfig.Installer,
	installerType appconfig.InstallerType,
	defaults *appconfig.AppConfigDefaults,
) *appconfig.Installer {
	if defaults != nil && *defaults.Type != nil {
		if val, ok := (*defaults.Type)[installerType]; ok {
			logger.Debug("Applying defaults for %s", installerType)
			if val.Opts != nil {
				installer.Opts = val.Opts
			}
			if val.Platforms != nil {
				installer.Platforms = val.Platforms
			}
			if val.PreUpdate != nil {
				installer.PreUpdate = val.PreUpdate
			}
			if val.PostUpdate != nil {
				installer.PostUpdate = val.PostUpdate
			}
			if val.PreInstall != nil {
				installer.PreInstall = val.PreInstall
			}
			if val.PostInstall != nil {
				installer.PostInstall = val.PostInstall
			}
			if val.CheckHasUpdate != nil {
				installer.CheckHasUpdate = val.CheckHasUpdate
			}
			if val.CheckInstalled != nil {
				installer.CheckInstalled = val.CheckInstalled
			}
		}
	}
	return installer
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
	info := installer.GetInfo()
	name := *info.Name
	logger.Debug("Checking if %s (%s) should run on %s", name, info.Type, GetCurrentPlatform())
	curOS := GetCurrentPlatform()
	env := config.Environ()
	if !GetShouldRunOnOS(installer, curOS) {
		logger.Debug("%s should not run on %s, skipping", name, curOS)
		return nil
	}
	logger.Debug("Checking if %s (%s) is installed", name, info.Type)
	err, installed := installer.CheckIsInstalled()
	if err != nil {
		return err
	}
	if installed {
		logger.Debug("%s (%s) is already installed", name, info.Type)
		if config.CheckUpdates {
			logger.Info("Checking updates for %s (%s)", name, info.Type)
			err, needsUpdate := installer.CheckNeedsUpdate()
			if err != nil {
				return err
			}
			if needsUpdate {
				logger.Info("%s (%s) has an update", name, info.Type)
				if info.PreUpdate != nil {
					logger.Debug("Running pre-update command for %s (%s)", name, info.Type)
					err := utils.RunCmdPassThrough(env, utils.GetOSShell(), utils.GetOSShellArgs(*info.PreUpdate)...)
					if err != nil {
						return err
					}
				}
				logger.Debug("Running update for %s (%s)", name, info.Type)
				installer.Update()
				if info.PostUpdate != nil {
					logger.Debug("Running post-update command for %s (%s)", name, info.Type)
					err := utils.RunCmdPassThrough(env, utils.GetOSShell(), utils.GetOSShellArgs(*info.PostUpdate)...)
					if err != nil {
						return err
					}
				}
			}
			return nil
		} else {
			return nil
		}
	}
	logger.Info("Installing %s (%s)", name, installer.GetInfo().Type)
	if info.PreInstall != nil {
		logger.Debug("Running pre-install command for %s (%s)", name, info.Type)
		err := utils.RunCmdPassThrough(env, utils.GetOSShell(), utils.GetOSShellArgs(*info.PreInstall)...)
		if err != nil {
			return err
		}
	}
	logger.Debug("Running installer for %s (%s)", name, info.Type)
	err = installer.Install()
	if info.PostInstall != nil {
		logger.Debug("Running post-install command for %s (%s)", name, info.Type)
		err := utils.RunCmdPassThrough(env, utils.GetOSShell(), utils.GetOSShellArgs(*info.PostInstall)...)
		if err != nil {
			return err
		}
	}
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
