package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
	"github.com/chenasraf/sofmani/utils"
)

type IInstaller interface {
	GetData() *appconfig.InstallerData
	CheckIsInstalled() (error, bool)
	CheckNeedsUpdate() (error, bool)
	Install() error
	Update() error
}

func GetInstaller(config *appconfig.AppConfig, data *appconfig.InstallerData) (error, IInstaller) {
	data = InstallerWithDefaults(data, data.Type, config.Defaults)
	switch data.Type {
	case appconfig.InstallerTypeGroup:
		return nil, NewGroupInstaller(config, data)
	case appconfig.InstallerTypeBrew:
		return nil, NewBrewInstaller(config, data)
	case appconfig.InstallerTypeShell:
		return nil, NewShellInstaller(config, data)
	case appconfig.InstallerTypeRsync:
		return nil, NewRsyncInstaller(config, data)
	case appconfig.InstallerTypeNpm, appconfig.InstallerTypePnpm, appconfig.InstallerTypeYarn:
		return nil, NewNpmInstaller(config, data)
	case appconfig.InstallerTypeApt, appconfig.InstallerTypeApk:
		return nil, NewAptInstaller(config, data)
	case appconfig.InstallerTypePipx:
		return nil, NewPipxInstaller(config, data)
	case appconfig.InstallerTypeGit:
		return nil, NewGitInstaller(config, data)
	case appconfig.InstallerTypeManifest:
		return nil, NewManifestInstaller(config, data)
	}
	return nil, nil
}

func RunInstaller(config *appconfig.AppConfig, installer IInstaller) error {
	info := installer.GetData()
	name := *info.Name
	curOS := platform.GetPlatform()
	logger.Debug("Checking if %s (%s) should run on %s", name, info.Type, curOS)
	env := config.Environ()
	if !installer.GetData().Platforms.GetShouldRunOnOS(curOS) {
		logger.Debug("%s should not run on %s, skipping", name, curOS)
		return nil
	}
	if !FilterInstaller(installer, config.Filter) {
		logger.Debug("%s is filtered, skipping", name)
		return nil
	}

	enabled, err := InstallerIsEnabled(installer)

	if err != nil {
		logger.Error("Failed to check if %s is enabled: %s", name, err)
		return nil
	}

	if !enabled {
		logger.Debug("%s is disabled, skipping", name)
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
				logger.Info("Updating %s (%s)", name, info.Type)
				if info.PreUpdate != nil {
					logger.Debug("Running pre-update command for %s (%s)", name, info.Type)
					err := utils.RunCmdPassThrough(env, utils.GetOSShell(installer.GetData().EnvShell), utils.GetOSShellArgs(*info.PreUpdate)...)
					if err != nil {
						return err
					}
				}
				logger.Debug("Running update command for %s (%s)", name, info.Type)
				installer.Update()
				if info.PostUpdate != nil {
					logger.Debug("Running post-update command for %s (%s)", name, info.Type)
					err := utils.RunCmdPassThrough(env, utils.GetOSShell(installer.GetData().EnvShell), utils.GetOSShellArgs(*info.PostUpdate)...)
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
	logger.Info("Installing %s (%s)", name, installer.GetData().Type)
	if info.PreInstall != nil {
		logger.Debug("Running pre-install command for %s (%s)", name, info.Type)
		err := utils.RunCmdPassThrough(env, utils.GetOSShell(installer.GetData().EnvShell), utils.GetOSShellArgs(*info.PreInstall)...)
		if err != nil {
			return err
		}
	}
	logger.Debug("Running installer for %s (%s)", name, info.Type)
	err = installer.Install()
	if info.PostInstall != nil {
		logger.Debug("Running post-install command for %s (%s)", name, info.Type)
		err := utils.RunCmdPassThrough(env, utils.GetOSShell(installer.GetData().EnvShell), utils.GetOSShellArgs(*info.PostInstall)...)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	return nil
}
