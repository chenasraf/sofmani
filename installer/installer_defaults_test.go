package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger.InitLogger(false)
}

func TestFillDefaults(t *testing.T) {
	t.Run("fills nil fields with empty values", func(t *testing.T) {
		data := &appconfig.InstallerData{}
		FillDefaults(data)

		assert.NotNil(t, data.Env)
		assert.NotNil(t, data.Opts)
		assert.NotNil(t, data.PlatformEnv)
		assert.NotNil(t, data.EnvShell)
		assert.NotNil(t, data.Platforms)
		assert.NotNil(t, data.Steps)
		assert.NotNil(t, data.Tags)
	})

	t.Run("does not overwrite existing values", func(t *testing.T) {
		existingEnv := map[string]string{"KEY": "VALUE"}
		existingOpts := map[string]any{"opt": "val"}
		data := &appconfig.InstallerData{
			Env:  &existingEnv,
			Opts: &existingOpts,
		}
		FillDefaults(data)

		assert.Equal(t, "VALUE", (*data.Env)["KEY"])
		assert.Equal(t, "val", (*data.Opts)["opt"])
	})

	t.Run("sets Linux-only platforms for apt installer", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Type: appconfig.InstallerTypeApt,
		}
		FillDefaults(data)

		assert.NotNil(t, data.Platforms.Only)
		assert.Contains(t, *data.Platforms.Only, platform.PlatformLinux)
	})

	t.Run("sets Linux-only platforms for apk installer", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Type: appconfig.InstallerTypeApk,
		}
		FillDefaults(data)

		assert.NotNil(t, data.Platforms.Only)
		assert.Contains(t, *data.Platforms.Only, platform.PlatformLinux)
	})

	t.Run("sets Linux-only platforms for pacman installer", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Type: appconfig.InstallerTypePacman,
		}
		FillDefaults(data)

		assert.NotNil(t, data.Platforms.Only)
		assert.Contains(t, *data.Platforms.Only, platform.PlatformLinux)
	})

	t.Run("sets Linux-only platforms for yay installer", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Type: appconfig.InstallerTypeYay,
		}
		FillDefaults(data)

		assert.NotNil(t, data.Platforms.Only)
		assert.Contains(t, *data.Platforms.Only, platform.PlatformLinux)
	})
}

func TestInstallerWithDefaults_Comprehensive(t *testing.T) {
	t.Run("applies base defaults when no type defaults", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Type: appconfig.InstallerTypeBrew,
		}
		result := InstallerWithDefaults(data, appconfig.InstallerTypeBrew, nil)

		assert.NotNil(t, result.Env)
		assert.NotNil(t, result.Opts)
	})

	t.Run("applies type-specific defaults", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Type: appconfig.InstallerTypeBrew,
		}
		defaultOpts := map[string]any{"tap": "default/tap"}
		defaults := &appconfig.AppConfigDefaults{
			Type: &map[appconfig.InstallerType]appconfig.InstallerData{
				appconfig.InstallerTypeBrew: {
					Opts: &defaultOpts,
				},
			},
		}
		result := InstallerWithDefaults(data, appconfig.InstallerTypeBrew, defaults)

		assert.Equal(t, "default/tap", (*result.Opts)["tap"])
	})

	t.Run("applies env defaults", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Type: appconfig.InstallerTypeBrew,
		}
		defaultEnv := map[string]string{"DEFAULT_VAR": "default_value"}
		defaults := &appconfig.AppConfigDefaults{
			Type: &map[appconfig.InstallerType]appconfig.InstallerData{
				appconfig.InstallerTypeBrew: {
					Env: &defaultEnv,
				},
			},
		}
		result := InstallerWithDefaults(data, appconfig.InstallerTypeBrew, defaults)

		assert.Equal(t, "default_value", (*result.Env)["DEFAULT_VAR"])
	})

	t.Run("applies hook defaults", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Type: appconfig.InstallerTypeBrew,
		}
		preInstall := "echo pre"
		postInstall := "echo post"
		preUpdate := "echo pre-update"
		postUpdate := "echo post-update"
		defaults := &appconfig.AppConfigDefaults{
			Type: &map[appconfig.InstallerType]appconfig.InstallerData{
				appconfig.InstallerTypeBrew: {
					PreInstall:  &preInstall,
					PostInstall: &postInstall,
					PreUpdate:   &preUpdate,
					PostUpdate:  &postUpdate,
				},
			},
		}
		result := InstallerWithDefaults(data, appconfig.InstallerTypeBrew, defaults)

		assert.Equal(t, "echo pre", *result.PreInstall)
		assert.Equal(t, "echo post", *result.PostInstall)
		assert.Equal(t, "echo pre-update", *result.PreUpdate)
		assert.Equal(t, "echo post-update", *result.PostUpdate)
	})

	t.Run("applies check command defaults", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Type: appconfig.InstallerTypeBrew,
		}
		checkInstalled := "which myapp"
		checkHasUpdate := "myapp --version"
		defaults := &appconfig.AppConfigDefaults{
			Type: &map[appconfig.InstallerType]appconfig.InstallerData{
				appconfig.InstallerTypeBrew: {
					CheckInstalled: &checkInstalled,
					CheckHasUpdate: &checkHasUpdate,
				},
			},
		}
		result := InstallerWithDefaults(data, appconfig.InstallerTypeBrew, defaults)

		assert.Equal(t, "which myapp", *result.CheckInstalled)
		assert.Equal(t, "myapp --version", *result.CheckHasUpdate)
	})

	t.Run("applies platform defaults", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Type: appconfig.InstallerTypeBrew,
		}
		platforms := platform.Platforms{
			Only: &[]platform.Platform{platform.PlatformMacos},
		}
		defaults := &appconfig.AppConfigDefaults{
			Type: &map[appconfig.InstallerType]appconfig.InstallerData{
				appconfig.InstallerTypeBrew: {
					Platforms: &platforms,
				},
			},
		}
		result := InstallerWithDefaults(data, appconfig.InstallerTypeBrew, defaults)

		assert.NotNil(t, result.Platforms.Only)
		assert.Contains(t, *result.Platforms.Only, platform.PlatformMacos)
	})

	t.Run("does not apply defaults for different installer type", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Type: appconfig.InstallerTypeBrew,
		}
		defaultOpts := map[string]any{"npm_option": "value"}
		defaults := &appconfig.AppConfigDefaults{
			Type: &map[appconfig.InstallerType]appconfig.InstallerData{
				appconfig.InstallerTypeNpm: {
					Opts: &defaultOpts,
				},
			},
		}
		result := InstallerWithDefaults(data, appconfig.InstallerTypeBrew, defaults)

		assert.NotContains(t, *result.Opts, "npm_option")
	})

	t.Run("handles nil defaults gracefully", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Type: appconfig.InstallerTypeBrew,
		}
		result := InstallerWithDefaults(data, appconfig.InstallerTypeBrew, nil)

		assert.NotNil(t, result)
		assert.NotNil(t, result.Opts)
	})

	t.Run("handles empty defaults type map", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Type: appconfig.InstallerTypeBrew,
		}
		defaults := &appconfig.AppConfigDefaults{
			Type: nil,
		}
		result := InstallerWithDefaults(data, appconfig.InstallerTypeBrew, defaults)

		assert.NotNil(t, result)
	})
}
