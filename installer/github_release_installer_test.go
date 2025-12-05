package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
	"github.com/stretchr/testify/assert"
)

func newTestGitHubReleaseInstaller(data *appconfig.InstallerData) *GitHubReleaseInstaller {
	return &GitHubReleaseInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Info: data,
	}
}

func TestGitHubReleaseValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid
	validData := &appconfig.InstallerData{
		Name: strPtr("ghr-valid"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"repository":        "owner/repo",
			"destination":       "/some/path",
			"download_filename": "file.tar.gz", // valid string
			"strategy":          "tar",
		},
	}
	assertNoValidationErrors(t, newTestGitHubReleaseInstaller(validData).Validate())

	// 🔴 Missing repository
	missingRepo := &appconfig.InstallerData{
		Name: strPtr("ghr-missing-repo"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"destination":       "/some/path",
			"download_filename": "file.tar.gz",
		},
	}
	assertValidationError(t, newTestGitHubReleaseInstaller(missingRepo).Validate(), "repository")

	// 🔴 Missing download_filename
	missingDownloadFilename := &appconfig.InstallerData{
		Name: strPtr("ghr-missing-download"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"repository":  "owner/repo",
			"destination": "/some/path",
		},
	}
	assertValidationError(t, newTestGitHubReleaseInstaller(missingDownloadFilename).Validate(), "download_filename")

	// 🔴 Empty per-platform download_filename
	emptyPlatformFilename := &appconfig.InstallerData{
		Name: strPtr("ghr-empty-platform-filename"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"repository":  "owner/repo",
			"destination": "/some/path",
			"download_filename": map[string]*string{
				string(platform.GetPlatform()): strPtr(""),
			},
		},
	}
	assertValidationError(t, newTestGitHubReleaseInstaller(emptyPlatformFilename).Validate(), "download_filename")

	// 🔴 Invalid strategy
	invalidStrategy := &appconfig.InstallerData{
		Name: strPtr("ghr-invalid-strategy"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"repository":        "owner/repo",
			"destination":       "/some/path",
			"download_filename": "file.tar.gz",
			"strategy":          "exe", // invalid
		},
	}
	assertValidationError(t, newTestGitHubReleaseInstaller(invalidStrategy).Validate(), "strategy")
}

func TestGitHubReleaseGetOpts(t *testing.T) {
	logger.InitLogger(false)

	t.Run("parses all options correctly", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("test-release"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"repository":        "owner/repo",
				"destination":       "/usr/local/bin",
				"download_filename": "app_{{ .Tag }}.tar.gz",
				"strategy":          "tar",
			},
		}
		installer := newTestGitHubReleaseInstaller(data)
		opts := installer.GetOpts()

		assert.Equal(t, "owner/repo", *opts.Repository)
		assert.Equal(t, "/usr/local/bin", *opts.Destination)
		assert.Equal(t, GitHubReleaseInstallStrategyTar, *opts.Strategy)
	})

	t.Run("handles nil opts", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("test-release"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: nil,
		}
		installer := newTestGitHubReleaseInstaller(data)
		opts := installer.GetOpts()

		assert.Nil(t, opts.Repository)
		assert.Nil(t, opts.Destination)
		assert.Nil(t, opts.Strategy)
	})

	t.Run("handles zip strategy", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("test-release"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"strategy": "zip",
			},
		}
		installer := newTestGitHubReleaseInstaller(data)
		opts := installer.GetOpts()

		assert.Equal(t, GitHubReleaseInstallStrategyZip, *opts.Strategy)
	})

	t.Run("handles none strategy", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("test-release"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"strategy": "none",
			},
		}
		installer := newTestGitHubReleaseInstaller(data)
		opts := installer.GetOpts()

		assert.Equal(t, GitHubReleaseInstallStrategyNone, *opts.Strategy)
	})
}

func TestGitHubReleaseGetBinName(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns bin_name when set", func(t *testing.T) {
		binName := "custom-bin"
		data := &appconfig.InstallerData{
			Name:    strPtr("my-app"),
			Type:    appconfig.InstallerTypeGitHubRelease,
			BinName: &binName,
		}
		installer := newTestGitHubReleaseInstaller(data)
		assert.Equal(t, "custom-bin", installer.GetBinName())
	})

	t.Run("returns base name when bin_name not set", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("owner/my-app"),
			Type: appconfig.InstallerTypeGitHubRelease,
		}
		installer := newTestGitHubReleaseInstaller(data)
		assert.Equal(t, "my-app", installer.GetBinName())
	})
}

func TestGitHubReleaseGetFilename(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns filename for current platform", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("test-release"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"download_filename": "app.tar.gz",
			},
		}
		installer := newTestGitHubReleaseInstaller(data)
		assert.Equal(t, "app.tar.gz", installer.GetFilename())
	})

	t.Run("returns empty string when not set", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("test-release"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{},
		}
		installer := newTestGitHubReleaseInstaller(data)
		assert.Equal(t, "", installer.GetFilename())
	})
}

func TestGitHubReleaseCacheOperations(t *testing.T) {
	logger.InitLogger(false)

	// Create a temporary cache directory for testing
	tmpDir, err := os.MkdirTemp("", "sofmani-test-cache")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// We need to test with actual cache operations
	// The cache uses utils.GetCacheDir() which we can't easily mock,
	// so we'll test the file operations directly

	t.Run("UpdateCache writes tag to file", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("test-cache-app"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"repository":        "owner/repo",
				"destination":       "/tmp",
				"download_filename": "app.tar.gz",
			},
		}
		installer := newTestGitHubReleaseInstaller(data)

		// Update the cache
		err := installer.UpdateCache("v1.0.0")
		assert.NoError(t, err)

		// Verify we can read it back
		cachedTag, err := installer.GetCachedTag()
		assert.NoError(t, err)
		assert.Equal(t, "v1.0.0", cachedTag)
	})

	t.Run("GetCachedTag returns empty for non-existent cache", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("non-existent-app-12345"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"repository":        "owner/repo",
				"destination":       "/tmp",
				"download_filename": "app.tar.gz",
			},
		}
		installer := newTestGitHubReleaseInstaller(data)

		cachedTag, err := installer.GetCachedTag()
		assert.NoError(t, err)
		assert.Equal(t, "", cachedTag)
	})

	t.Run("UpdateCache overwrites existing cache", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("test-overwrite-app"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"repository":        "owner/repo",
				"destination":       "/tmp",
				"download_filename": "app.tar.gz",
			},
		}
		installer := newTestGitHubReleaseInstaller(data)

		// Write initial version
		err := installer.UpdateCache("v1.0.0")
		assert.NoError(t, err)

		// Overwrite with new version
		err = installer.UpdateCache("v2.0.0")
		assert.NoError(t, err)

		// Verify new version
		cachedTag, err := installer.GetCachedTag()
		assert.NoError(t, err)
		assert.Equal(t, "v2.0.0", cachedTag)
	})
}

func TestGitHubReleaseGetDestination(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns destination from opts", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("test-release"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"destination": "/usr/local/bin",
			},
		}
		installer := newTestGitHubReleaseInstaller(data)
		assert.Equal(t, "/usr/local/bin", installer.GetDestination())
	})

	t.Run("returns current directory when destination not set", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("test-release"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{},
		}
		installer := newTestGitHubReleaseInstaller(data)
		wd, _ := os.Getwd()
		assert.Equal(t, wd, installer.GetDestination())
	})
}

func TestGitHubReleaseGetInstallDir(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns same as destination", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("test-release"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"destination": "/opt/bin",
			},
		}
		installer := newTestGitHubReleaseInstaller(data)
		assert.Equal(t, installer.GetDestination(), installer.GetInstallDir())
	})
}

func TestGitHubReleaseCheckIsInstalled(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns true when file exists", func(t *testing.T) {
		// Create a temp directory with the binary
		tmpDir, err := os.MkdirTemp("", "sofmani-install-test")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		// Create a fake binary
		binPath := filepath.Join(tmpDir, "myapp")
		err = os.WriteFile(binPath, []byte("fake binary"), 0755)
		assert.NoError(t, err)

		data := &appconfig.InstallerData{
			Name: strPtr("myapp"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"destination": tmpDir,
			},
		}
		installer := newTestGitHubReleaseInstaller(data)

		installed, err := installer.CheckIsInstalled()
		assert.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("returns false when file does not exist", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "sofmani-install-test")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		data := &appconfig.InstallerData{
			Name: strPtr("nonexistent-app"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"destination": tmpDir,
			},
		}
		installer := newTestGitHubReleaseInstaller(data)

		installed, err := installer.CheckIsInstalled()
		assert.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("uses custom check when provided", func(t *testing.T) {
		checkCmd := "true"
		data := &appconfig.InstallerData{
			Name:           strPtr("myapp"),
			Type:           appconfig.InstallerTypeGitHubRelease,
			CheckInstalled: &checkCmd,
		}
		installer := newTestGitHubReleaseInstaller(data)

		installed, err := installer.CheckIsInstalled()
		assert.NoError(t, err)
		assert.True(t, installed)
	})
}

func TestGitHubReleaseCheckNeedsUpdate(t *testing.T) {
	logger.InitLogger(false)

	t.Run("uses custom check when provided", func(t *testing.T) {
		checkCmd := "true" // returns success = update needed
		data := &appconfig.InstallerData{
			Name:           strPtr("myapp"),
			Type:           appconfig.InstallerTypeGitHubRelease,
			CheckHasUpdate: &checkCmd,
		}
		installer := newTestGitHubReleaseInstaller(data)

		needsUpdate, err := installer.CheckNeedsUpdate()
		assert.NoError(t, err)
		assert.True(t, needsUpdate)
	})

	t.Run("returns true when no cached tag", func(t *testing.T) {
		// Use a unique name that won't have a cache file
		data := &appconfig.InstallerData{
			Name: strPtr("unique-no-cache-app-99999"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"repository":        "owner/repo",
				"destination":       "/tmp",
				"download_filename": "app.tar.gz",
			},
		}
		installer := newTestGitHubReleaseInstaller(data)

		needsUpdate, err := installer.CheckNeedsUpdate()
		assert.NoError(t, err)
		assert.True(t, needsUpdate)
	})
}

func TestNewGitHubReleaseInstaller(t *testing.T) {
	logger.InitLogger(false)

	t.Run("creates installer with config and data", func(t *testing.T) {
		cfg := &appconfig.AppConfig{}
		data := &appconfig.InstallerData{
			Name: strPtr("test-release"),
			Type: appconfig.InstallerTypeGitHubRelease,
		}
		installer := NewGitHubReleaseInstaller(cfg, data)

		assert.NotNil(t, installer)
		assert.Equal(t, cfg, installer.Config)
		assert.Equal(t, data, installer.Info)
		assert.Equal(t, data, installer.Data)
	})
}
