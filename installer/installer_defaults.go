package installer

import (
	"maps"
	"slices"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/machine"
	"github.com/chenasraf/sofmani/platform"
)

// platformLockedTypes maps installer types to the platforms their package manager exists on.
// The restriction is not configurable: honoring a `platforms` value for one of these would run
// the step somewhere its package manager cannot be there.
var platformLockedTypes = map[appconfig.InstallerType][]platform.Platform{
	appconfig.InstallerTypeApt:    {platform.PlatformLinux},
	appconfig.InstallerTypeApk:    {platform.PlatformLinux},
	appconfig.InstallerTypePacman: {platform.PlatformLinux},
	appconfig.InstallerTypeYay:    {platform.PlatformLinux},
}

// lockPlatforms pins the platforms of a platform-locked type, discarding any `platforms` the
// manifest or the type defaults set. Use `enabled` to turn such a step off.
func lockPlatforms(data *appconfig.InstallerData) {
	locked, ok := platformLockedTypes[data.Type]
	if !ok {
		return
	}
	if data.Platforms == nil {
		data.Platforms = &platform.Platforms{}
	}
	configured := data.Platforms.Except != nil ||
		(data.Platforms.Only != nil && !slices.Equal(*data.Platforms.Only, locked))
	if configured {
		logger.Debug("Ignoring platforms for %s: the type only runs on %s", data.Type, locked)
	}
	only := slices.Clone(locked)
	data.Platforms.Only = &only
	data.Platforms.Except = nil
}

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
			if override.Machines != nil {
				data.Machines = override.Machines
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
			if override.Verbose != nil && data.Verbose == nil {
				data.Verbose = override.Verbose
			}
			if override.AllowFailure != nil && data.AllowFailure == nil {
				data.AllowFailure = override.AllowFailure
			}
			if override.ConfirmInstall != nil && data.ConfirmInstall == nil {
				data.ConfirmInstall = override.ConfirmInstall
			}
			if override.ConfirmUpdate != nil && data.ConfirmUpdate == nil {
				data.ConfirmUpdate = override.ConfirmUpdate
			}
		}
	}
	// The type defaults may have replaced the platforms wholesale, so pin them once more.
	lockPlatforms(data)
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
	if data.Machines == nil {
		machines := machine.Machines{}
		data.Machines = &machines
	}
	if data.Steps == nil {
		data.Steps = &[]appconfig.InstallerData{}
	}
	if data.Tags == nil {
		str := ""
		data.Tags = &str
	}
	lockPlatforms(data)
}
