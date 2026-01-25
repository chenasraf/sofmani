package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
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
		Name: strPtr("manifest-valid"),
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
		Name: strPtr("manifest-missing-source"),
		Type: appconfig.InstallerTypeManifest,
		Opts: &map[string]any{
			"path": "some/path",
		},
	}
	assertValidationError(t, newTestManifestInstaller(missingSource).Validate(), "source")

	// 🔴 Missing path
	missingPath := &appconfig.InstallerData{
		Name: strPtr("manifest-missing-path"),
		Type: appconfig.InstallerTypeManifest,
		Opts: &map[string]any{
			"source": "https://example.com/repo.git",
		},
	}
	assertValidationError(t, newTestManifestInstaller(missingPath).Validate(), "path")

	// 🔴 Empty ref (not nil, just empty)
	emptyRef := &appconfig.InstallerData{
		Name: strPtr("manifest-empty-ref"),
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
		Name: strPtr("manifest-nil-opts"),
		Type: appconfig.InstallerTypeManifest,
		Opts: nil,
	}
	assertValidationError(t, newTestManifestInstaller(nilOpts).Validate(), "source")
}

func TestManifestGetOpts(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns all opts when set", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("manifest-test"),
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
			Name: strPtr("manifest-test"),
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
			Name: strPtr("manifest-test"),
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
			Name: strPtr("manifest-test"),
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
			Name: strPtr("manifest-test"),
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
			Name: strPtr("manifest-test"),
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
			Name:           strPtr("manifest-test"),
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
			Name: strPtr("manifest-test"),
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
			Name:           strPtr("manifest-test"),
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
			Name: strPtr("manifest-test"),
			Type: appconfig.InstallerTypeManifest,
		}
		installer := NewManifestInstaller(cfg, data)

		assert.NotNil(t, installer)
		assert.Equal(t, cfg, installer.Config)
		assert.Equal(t, data, installer.Info)
		assert.Equal(t, data, installer.Data)
	})
}
