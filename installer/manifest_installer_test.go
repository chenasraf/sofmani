package installer

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func newTestManifestInstaller(data *appconfig.InstallerData) *ManifestInstaller {
	return &ManifestInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config: nil,
		Info:   data,
	}
}

func TestManifestValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid
	validData := &appconfig.InstallerData{
		Name: lo.ToPtr("manifest-valid"),
		Type: appconfig.InstallerTypeManifest,
		Opts: &map[string]any{
			"source": "https://example.com/repo.git",
			"path":   "manifests/installer.yml",
			"ref":    "main",
		},
	}
	assertNoValidationErrors(t, newTestManifestInstaller(validData).Validate())

	// 🔴 Missing source
	missingSource := &appconfig.InstallerData{
		Name: lo.ToPtr("manifest-missing-source"),
		Type: appconfig.InstallerTypeManifest,
		Opts: &map[string]any{
			"path": "some/path",
		},
	}
	assertValidationError(t, newTestManifestInstaller(missingSource).Validate(), "source")

	// 🔴 Missing path
	missingPath := &appconfig.InstallerData{
		Name: lo.ToPtr("manifest-missing-path"),
		Type: appconfig.InstallerTypeManifest,
		Opts: &map[string]any{
			"source": "https://example.com/repo.git",
		},
	}
	assertValidationError(t, newTestManifestInstaller(missingPath).Validate(), "path")

	// 🔴 Empty ref (not nil, just empty)
	emptyRef := &appconfig.InstallerData{
		Name: lo.ToPtr("manifest-empty-ref"),
		Type: appconfig.InstallerTypeManifest,
		Opts: &map[string]any{
			"source": "https://example.com/repo.git",
			"path":   "install.yml",
			"ref":    "",
		},
	}
	assertValidationError(t, newTestManifestInstaller(emptyRef).Validate(), "ref")

	// 🔴 Nil opts
	nilOpts := &appconfig.InstallerData{
		Name: lo.ToPtr("manifest-nil-opts"),
		Type: appconfig.InstallerTypeManifest,
		Opts: nil,
	}
	assertValidationError(t, newTestManifestInstaller(nilOpts).Validate(), "source")
}

func TestManifestGetOpts(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns all opts when set", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("manifest-test"),
			Type: appconfig.InstallerTypeManifest,
			Opts: &map[string]any{
				"source": "https://github.com/user/repo.git",
				"path":   "manifest.yml",
				"ref":    "develop",
			},
		}
		installer := newTestManifestInstaller(data)
		opts := installer.GetOpts()

		assert.NotNil(t, opts.Source)
		assert.Equal(t, "https://github.com/user/repo.git", *opts.Source)
		assert.NotNil(t, opts.Path)
		assert.Equal(t, "manifest.yml", *opts.Path)
		assert.NotNil(t, opts.Ref)
		assert.Equal(t, "develop", *opts.Ref)
	})

	t.Run("returns nil fields when opts is nil", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("manifest-test"),
			Type: appconfig.InstallerTypeManifest,
			Opts: nil,
		}
		installer := newTestManifestInstaller(data)
		opts := installer.GetOpts()

		assert.Nil(t, opts.Source)
		assert.Nil(t, opts.Path)
		assert.Nil(t, opts.Ref)
	})

	t.Run("handles partial opts", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("manifest-test"),
			Type: appconfig.InstallerTypeManifest,
			Opts: &map[string]any{
				"source": "https://github.com/user/repo.git",
			},
		}
		installer := newTestManifestInstaller(data)
		opts := installer.GetOpts()

		assert.NotNil(t, opts.Source)
		assert.Equal(t, "https://github.com/user/repo.git", *opts.Source)
		assert.Nil(t, opts.Path)
		assert.Nil(t, opts.Ref)
	})

	t.Run("handles wrong type values gracefully", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("manifest-test"),
			Type: appconfig.InstallerTypeManifest,
			Opts: &map[string]any{
				"source": 123,     // Wrong type
				"path":   true,    // Wrong type
				"ref":    []int{}, // Wrong type
			},
		}
		installer := newTestManifestInstaller(data)
		opts := installer.GetOpts()

		// Should return nil when type assertion fails
		assert.Nil(t, opts.Source)
		assert.Nil(t, opts.Path)
		assert.Nil(t, opts.Ref)
	})
}

func TestManifestGetData(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns the installer data", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("manifest-test"),
			Type: appconfig.InstallerTypeManifest,
		}
		installer := newTestManifestInstaller(data)
		result := installer.GetData()

		assert.Equal(t, data, result)
		assert.Equal(t, "manifest-test", *result.Name)
	})
}

func TestManifestCheckIsInstalled(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns false when no custom check", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("manifest-test"),
			Type: appconfig.InstallerTypeManifest,
		}
		installer := newTestManifestInstaller(data)
		result, err := installer.CheckIsInstalled()

		assert.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("runs custom check when provided", func(t *testing.T) {
		checkCmd := "true"
		data := &appconfig.InstallerData{
			Name:           lo.ToPtr("manifest-test"),
			Type:           appconfig.InstallerTypeManifest,
			CheckInstalled: &checkCmd,
		}
		installer := newTestManifestInstaller(data)
		result, err := installer.CheckIsInstalled()

		assert.NoError(t, err)
		assert.True(t, result)
	})
}

func TestManifestCheckNeedsUpdate(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns true when no custom check", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("manifest-test"),
			Type: appconfig.InstallerTypeManifest,
		}
		installer := newTestManifestInstaller(data)
		result, err := installer.CheckNeedsUpdate()

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("runs custom check when provided", func(t *testing.T) {
		checkCmd := "false" // Returns exit code 1, meaning no update
		data := &appconfig.InstallerData{
			Name:           lo.ToPtr("manifest-test"),
			Type:           appconfig.InstallerTypeManifest,
			CheckHasUpdate: &checkCmd,
		}
		installer := newTestManifestInstaller(data)
		result, err := installer.CheckNeedsUpdate()

		assert.NoError(t, err)
		assert.False(t, result)
	})
}

func TestNewManifestInstaller(t *testing.T) {
	logger.InitLogger(false)

	t.Run("creates installer with config and data", func(t *testing.T) {
		cfg := &appconfig.AppConfig{}
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("manifest-test"),
			Type: appconfig.InstallerTypeManifest,
		}
		installer := NewManifestInstaller(cfg, data)

		assert.NotNil(t, installer)
		assert.Equal(t, cfg, installer.Config)
		assert.Equal(t, data, installer.Info)
		assert.Equal(t, data, installer.Data)
	})
}

func manifestData(opts map[string]any) *appconfig.InstallerData {
	base := map[string]any{"source": "/tmp", "path": "manifest.yml"}
	maps.Copy(base, opts)
	return &appconfig.InstallerData{
		Name: lo.ToPtr("manifest-test"),
		Type: appconfig.InstallerTypeManifest,
		Opts: &base,
	}
}

func TestManifestInheritParsing(t *testing.T) {
	logger.InitLogger(false)

	t.Run("inherits nothing by default", func(t *testing.T) {
		opts := newTestManifestInstaller(manifestData(nil)).GetOpts()
		assert.Equal(t, ManifestInherit{}, opts.Inherit)
	})

	t.Run("true inherits everything", func(t *testing.T) {
		opts := newTestManifestInstaller(manifestData(map[string]any{"inherit": true})).GetOpts()
		assert.Equal(t, inheritAll, opts.Inherit)
	})

	t.Run("map inherits only the listed settings", func(t *testing.T) {
		data := manifestData(map[string]any{"inherit": map[string]any{"defaults": true, "env": false}})
		opts := newTestManifestInstaller(data).GetOpts()
		assert.Equal(t, ManifestInherit{Defaults: true}, opts.Inherit)
	})

	t.Run("unknown setting fails validation", func(t *testing.T) {
		data := manifestData(map[string]any{"inherit": map[string]any{"envs": true}})
		assertValidationError(t, newTestManifestInstaller(data).Validate(), "inherit")
	})

	t.Run("non-bool non-map fails validation", func(t *testing.T) {
		data := manifestData(map[string]any{"inherit": "yes"})
		assertValidationError(t, newTestManifestInstaller(data).Validate(), "inherit")
	})
}

func TestManifestOverridesParsing(t *testing.T) {
	logger.InitLogger(false)

	t.Run("parses global settings", func(t *testing.T) {
		data := manifestData(map[string]any{"overrides": map[string]any{
			"env":             map[string]any{"FOO": "bar"},
			"platform_env":    map[string]any{"linux": map[string]any{"BAZ": "qux"}},
			"repo_update":     map[string]any{"brew": "never"},
			"check_updates":   true,
			"machine_aliases": map[string]any{"work": "abc"},
			"defaults": map[string]any{"type": map[string]any{
				"brew": map[string]any{"opts": map[string]any{"tap": "x/y"}},
			}},
		}})
		inst := newTestManifestInstaller(data)
		assertNoValidationErrors(t, inst.Validate())
		o := inst.GetOpts().Overrides
		assert.Equal(t, "bar", (*o.Env)["FOO"])
		assert.Equal(t, "qux", (*o.PlatformEnv.Linux)["BAZ"])
		assert.Equal(t, appconfig.RepoUpdateNever, (*o.RepoUpdate)[appconfig.InstallerTypeBrew])
		assert.True(t, *o.CheckUpdates)
		assert.Equal(t, "abc", (*o.MachineAliases)["work"])
		assert.Equal(t, "x/y", (*(*o.Defaults.Type)[appconfig.InstallerTypeBrew].Opts)["tap"])
	})

	t.Run("unknown setting fails validation", func(t *testing.T) {
		data := manifestData(map[string]any{"overrides": map[string]any{"install": []any{}}})
		assertValidationError(t, newTestManifestInstaller(data).Validate(), "overrides")
	})

	t.Run("invalid overrides refuse to fetch", func(t *testing.T) {
		data := manifestData(map[string]any{"overrides": "nope"})
		assert.Error(t, newTestManifestInstaller(data).FetchManifest())
	})
}

func TestManifestBuildConfig(t *testing.T) {
	logger.InitLogger(false)

	parent := func() *appconfig.AppConfig {
		return &appconfig.AppConfig{
			CheckUpdates:   lo.ToPtr(true),
			Env:            &map[string]string{"SECRET": "parent", "SHARED": "parent"},
			RepoUpdate:     &map[appconfig.InstallerType]appconfig.RepoUpdateMode{appconfig.InstallerTypeBrew: appconfig.RepoUpdateNever},
			MachineAliases: &map[string]string{"home": "123"},
			Defaults: &appconfig.AppConfigDefaults{Type: &map[appconfig.InstallerType]appconfig.InstallerData{
				appconfig.InstallerTypeShell: {Verbose: lo.ToPtr(true)},
			}},
		}
	}
	child := func() *appconfig.AppConfig {
		return &appconfig.AppConfig{Env: &map[string]string{"SHARED": "child", "OWN": "child"}}
	}
	build := func(p *appconfig.AppConfig, opts map[string]any) *appconfig.AppConfig {
		inst := NewManifestInstaller(p, manifestData(opts))
		return inst.buildManifestConfig(child(), inst.GetOpts())
	}

	t.Run("inherits nothing by default", func(t *testing.T) {
		cfg := build(parent(), nil)
		assert.Equal(t, map[string]string{"SHARED": "child", "OWN": "child"}, *cfg.Env)
		assert.False(t, *cfg.CheckUpdates)
		assert.Nil(t, cfg.RepoUpdate)
		assert.Nil(t, cfg.MachineAliases)
		assert.Nil(t, cfg.Defaults)
	})

	t.Run("inherited settings win over the manifest's own", func(t *testing.T) {
		cfg := build(parent(), map[string]any{"inherit": true})
		assert.Equal(t, map[string]string{"SECRET": "parent", "SHARED": "parent", "OWN": "child"}, *cfg.Env)
		assert.True(t, *cfg.CheckUpdates)
		assert.Equal(t, appconfig.RepoUpdateNever, cfg.GetRepoUpdateMode(appconfig.InstallerTypeBrew))
		assert.Equal(t, "123", (*cfg.MachineAliases)["home"])
		assert.Contains(t, *cfg.Defaults.Type, appconfig.InstallerTypeShell)
	})

	t.Run("overrides win over everything", func(t *testing.T) {
		cfg := build(parent(), map[string]any{
			"inherit":   map[string]any{"env": true},
			"overrides": map[string]any{"env": map[string]any{"SHARED": "override"}, "check_updates": true},
		})
		assert.Equal(t, map[string]string{"SECRET": "parent", "SHARED": "override", "OWN": "child"}, *cfg.Env)
		assert.True(t, *cfg.CheckUpdates)
	})

	t.Run("never writes into the parent config", func(t *testing.T) {
		p := parent()
		build(p, map[string]any{"inherit": true, "overrides": map[string]any{"env": map[string]any{"X": "y"}}})
		assert.Equal(t, map[string]string{"SECRET": "parent", "SHARED": "parent"}, *p.Env)
	})
}

func TestManifestEnvIsolation(t *testing.T) {
	logger.InitLogger(false)

	dir := t.TempDir()
	out := filepath.Join(dir, "out.txt")
	manifest := fmt.Sprintf(`install:
  - name: probe
    type: shell
    opts:
      command: 'printf "%%s|%%s" "$SOFMANI_TEST_SECRET" "$SOFMANI_TEST_OWN" > %s'
`, out)
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "manifest.yml"), []byte(manifest), 0o644))

	run := func(opts map[string]any) string {
		// main applies the root env to the process the same way.
		t.Setenv("SOFMANI_TEST_SECRET", "hunter2")
		parent := &appconfig.AppConfig{
			CheckUpdates: lo.ToPtr(false),
			Env:          &map[string]string{"SOFMANI_TEST_SECRET": "hunter2"},
		}
		base := map[string]any{"source": dir, "path": "manifest.yml"}
		maps.Copy(base, opts)
		data := &appconfig.InstallerData{Name: lo.ToPtr("isolated"), Type: appconfig.InstallerTypeManifest, Opts: &base}
		assert.NoError(t, NewManifestInstaller(parent, data).Install())
		assert.Equal(t, "hunter2", os.Getenv("SOFMANI_TEST_SECRET"), "parent env is restored afterwards")
		_, leaked := os.LookupEnv("SOFMANI_TEST_OWN")
		assert.False(t, leaked, "manifest env does not outlive the manifest")
		content, err := os.ReadFile(out)
		assert.NoError(t, err)
		return string(content)
	}

	overrides := map[string]any{"env": map[string]any{"SOFMANI_TEST_OWN": "mine"}}
	assert.Equal(t, "|mine", run(map[string]any{"overrides": overrides}))
	assert.Equal(t, "hunter2|mine", run(map[string]any{"overrides": overrides, "inherit": map[string]any{"env": true}}))
}
