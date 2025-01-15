package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
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
	case appconfig.InstallerTypeApt:
		return nil, NewAptInstaller(config, installer)
	case appconfig.InstallerTypeGit:
		return nil, NewGitInstaller(config, installer)
	case appconfig.InstallerTypeManifest:
		return nil, NewManifestInstaller(config, installer)
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
				o := *val.Opts
				o2 := *installer.Opts
				for k, v := range o {
					o2[k] = v
				}
			}
			if val.EnvShell != nil {
				installer.EnvShell = val.EnvShell
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

func RunInstaller(config *appconfig.AppConfig, installer IInstaller) error {
	info := installer.GetInfo()
	name := *info.Name
	curOS := platform.GetPlatform()
	logger.Debug("Checking if %s (%s) should run on %s", name, info.Type, curOS)
	env := config.Environ()
	if !installer.GetInfo().Platforms.GetShouldRunOnOS(curOS) {
		logger.Debug("%s should not run on %s, skipping", name, curOS)
		return nil
	}
	if !FilterIsMatch(config.Filter, name) {
		logger.Debug("%s is filtered, skipping", name)
		return nil
	}
	logger.Debug("Checking %s (%s)", name, info.Type)
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
					err := utils.RunCmdPassThrough(env, utils.GetOSShell(installer.GetInfo().EnvShell), utils.GetOSShellArgs(*info.PreUpdate)...)
					if err != nil {
						return err
					}
				}
				logger.Debug("Running update for %s (%s)", name, info.Type)
				installer.Update()
				if info.PostUpdate != nil {
					logger.Debug("Running post-update command for %s (%s)", name, info.Type)
					err := utils.RunCmdPassThrough(env, utils.GetOSShell(installer.GetInfo().EnvShell), utils.GetOSShellArgs(*info.PostUpdate)...)
					if err != nil {
						return err
					}
				}
			} else {
				logger.Info("%s (%s) is up-to-date", name, info.Type)
			}
			return nil
		} else {
			return nil
		}
	}
	logger.Info("Installing %s (%s)", name, installer.GetInfo().Type)
	if info.PreInstall != nil {
		logger.Debug("Running pre-install command for %s (%s)", name, info.Type)
		err := utils.RunCmdPassThrough(env, utils.GetOSShell(installer.GetInfo().EnvShell), utils.GetOSShellArgs(*info.PreInstall)...)
		if err != nil {
			return err
		}
	}
	logger.Debug("Running installer for %s (%s)", name, info.Type)
	err = installer.Install()
	if info.PostInstall != nil {
		logger.Debug("Running post-install command for %s (%s)", name, info.Type)
		err := utils.RunCmdPassThrough(env, utils.GetOSShell(installer.GetInfo().EnvShell), utils.GetOSShellArgs(*info.PostInstall)...)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	return nil
}
