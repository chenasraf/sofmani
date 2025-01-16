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
	case appconfig.InstallerTypeApt:
		return nil, NewAptInstaller(config, data)
	case appconfig.InstallerTypeGit:
		return nil, NewGitInstaller(config, data)
	case appconfig.InstallerTypeManifest:
		return nil, NewManifestInstaller(config, data)
	}
	return nil, nil
}

func InstallerWithDefaults(
	data *appconfig.InstallerData,
	installerType appconfig.InstallerType,
	defaults *appconfig.AppConfigDefaults,
) *appconfig.InstallerData {
	// set base defaults
	FillDefaults(data)

	// per-type overrides from defaults
	if defaults != nil && *defaults.Type != nil {
		if override, ok := (*defaults.Type)[installerType]; ok {
			logger.Debug("Applying defaults for %s", installerType)
			if override.Opts != nil {
				source := *override.Opts
				target := *data.Opts
				for k, v := range source {
					target[k] = v
				}
			}
			if override.Env != nil {
				source := *override.Env
				target := *data.Env
				for k, v := range source {
					target[k] = v
				}
			}
			if override.PlatformEnv != nil {
				source := *override.PlatformEnv
				targetBase := *data.PlatformEnv
				target := *targetBase.MacOS
				for k, v := range *source.MacOS {
					target[k] = v
				}
				target = *targetBase.Linux
				for k, v := range *source.Linux {
					target[k] = v
				}
				target = *targetBase.Windows
				for k, v := range *source.Windows {
					target[k] = v
				}
			}
			if override.EnvShell != nil {
				// TODO override key by key
				data.EnvShell = override.EnvShell
			}
			if override.Platforms != nil {
				data.Platforms = override.Platforms
			}
			if override.PreUpdate != nil {
				data.PreUpdate = override.PreUpdate
			}
			if override.PostUpdate != nil {
				data.PostUpdate = override.PostUpdate
			}
			if override.PreInstall != nil {
				data.PreInstall = override.PreInstall
			}
			if override.PostInstall != nil {
				data.PostInstall = override.PostInstall
			}
			if override.CheckHasUpdate != nil {
				data.CheckHasUpdate = override.CheckHasUpdate
			}
			if override.CheckInstalled != nil {
				data.CheckInstalled = override.CheckInstalled
			}
		}
	}
	return data
}

func FillDefaults(data *appconfig.InstallerData) {
	if data.Env == nil {
		data.Env = &map[string]string{}
	}
	if data.Opts == nil {
		data.Opts = &map[string]any{}
	}
	if data.PlatformEnv == nil {
		env := platform.PlatformMap[map[string]string]{
			MacOS:   &map[string]string{},
			Linux:   &map[string]string{},
			Windows: &map[string]string{},
		}
		data.PlatformEnv = &env
	}
	if data.Platforms == nil {
		platforms := platform.Platforms{}
		data.Platforms = &platforms
	}
	if data.Steps == nil {
		data.Steps = &[]appconfig.InstallerData{}
	}
	if data.Tags == nil {
		str := ""
		data.Tags = &str
	}
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
