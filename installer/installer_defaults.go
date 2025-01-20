package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
)

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
				source := *override.EnvShell
				target := *data.EnvShell
				if source.MacOS != nil {
					*target.MacOS = *source.MacOS
				}
				if source.Linux != nil {
					*target.Linux = *source.Linux
				}
				if source.Windows != nil {
					*target.Windows = *source.Windows
				}
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
	// Default overrides per type
	switch data.Type {
	case appconfig.InstallerTypeApt, appconfig.InstallerTypeApk:
		data.Platforms = &platform.Platforms{
			Only: &[]platform.Platform{platform.PlatformLinux},
		}
	}
}
