package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
	"maps"
)

// InstallerWithDefaults applies default configurations to an installer data object.
// It first applies base defaults using FillDefaults, and then applies type-specific defaults.
func InstallerWithDefaults(
	data *appconfig.InstallerData,
	installerType appconfig.InstallerType,
	defaults *appconfig.AppConfigDefaults,
) *appconfig.InstallerData {
	// set base defaults
	FillDefaults(data)

	// per-type overrides from defaults
	if defaults != nil && defaults.Type != nil {
		if override, ok := (*defaults.Type)[installerType]; ok {
			logger.Debug("Applying defaults for %s", installerType)
			if override.Opts != nil {
				source := *override.Opts
				target := *data.Opts
				maps.Copy(target, source)
			}
			if override.Env != nil {
				source := *override.Env
				target := *data.Env
				maps.Copy(target, source)
			}
			if override.PlatformEnv != nil {
				source := *override.PlatformEnv
				targetBase := *data.PlatformEnv
				if source.MacOS != nil && targetBase.MacOS != nil {
					target := *targetBase.MacOS
					maps.Copy(target, *source.MacOS)
				}
				if source.Linux != nil && targetBase.Linux != nil {
					target := *targetBase.Linux
					maps.Copy(target, *source.Linux)
				}
				if source.Windows != nil && targetBase.Windows != nil {
					target := *targetBase.Windows
					maps.Copy(target, *source.Windows)
				}
			}
			if override.EnvShell != nil {
				source := *override.EnvShell
				target := data.EnvShell // data.EnvShell should be initialized by FillDefaults
				if target == nil {      // Should not happen if FillDefaults is called
					data.EnvShell = &platform.PlatformMap[string]{}
					target = data.EnvShell
				}
				if source.MacOS != nil {
					target.MacOS = source.MacOS
				}
				if source.Linux != nil {
					target.Linux = source.Linux
				}
				if source.Windows != nil {
					target.Windows = source.Windows
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

// FillDefaults initializes nil fields in an InstallerData object with empty values.
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
	if data.EnvShell == nil { // Added default for EnvShell
		shell := platform.PlatformMap[string]{}
		data.EnvShell = &shell
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
	case appconfig.InstallerTypePacman, appconfig.InstallerTypeYay:
		data.Platforms = &platform.Platforms{
			Only: &[]platform.Platform{platform.PlatformLinux},
		}
	}
}
