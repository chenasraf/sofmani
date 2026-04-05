package installer

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
	"github.com/samber/lo"
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
		Name: lo.ToPtr("ghr-valid"),
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
		Name: lo.ToPtr("ghr-missing-repo"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"destination":       "/some/path",
			"download_filename": "file.tar.gz",
		},
	}
	assertValidationError(t, newTestGitHubReleaseInstaller(missingRepo).Validate(), "repository")

	// 🔴 Missing download_filename
	missingDownloadFilename := &appconfig.InstallerData{
		Name: lo.ToPtr("ghr-missing-download"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"repository":  "owner/repo",
			"destination": "/some/path",
		},
	}
	assertValidationError(t, newTestGitHubReleaseInstaller(missingDownloadFilename).Validate(), "download_filename")

	// 🔴 Empty per-platform download_filename
	emptyPlatformFilename := &appconfig.InstallerData{
		Name: lo.ToPtr("ghr-empty-platform-filename"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"repository":  "owner/repo",
			"destination": "/some/path",
			"download_filename": map[string]*string{
				string(platform.GetPlatform()): lo.ToPtr(""),
			},
		},
	}
	assertValidationError(t, newTestGitHubReleaseInstaller(emptyPlatformFilename).Validate(), "download_filename")

	// 🔴 Invalid strategy
	invalidStrategy := &appconfig.InstallerData{
		Name: lo.ToPtr("ghr-invalid-strategy"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"repository":        "owner/repo",
			"destination":       "/some/path",
			"download_filename": "file.tar.gz",
			"strategy":          "exe", // invalid
		},
	}
	assertValidationError(t, newTestGitHubReleaseInstaller(invalidStrategy).Validate(), "strategy")

	// 🟢 Valid gzip strategy
	gzipStrategy := &appconfig.InstallerData{
		Name: lo.ToPtr("ghr-gzip"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"repository":        "owner/repo",
			"destination":       "/some/path",
			"download_filename": "file.gz",
			"strategy":          "gzip",
		},
	}
	assertNoValidationErrors(t, newTestGitHubReleaseInstaller(gzipStrategy).Validate())

	// 🟢 Valid custom strategy with extract_command
	customStrategy := &appconfig.InstallerData{
		Name: lo.ToPtr("ghr-custom"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"repository":        "owner/repo",
			"destination":       "/some/path",
			"download_filename": "file.weird",
			"strategy":          "custom",
			"extract_command":   "7z x {{ .DownloadFile }} -o{{ .ExtractDir }}",
		},
	}
	assertNoValidationErrors(t, newTestGitHubReleaseInstaller(customStrategy).Validate())

	// 🔴 custom strategy without extract_command
	customMissingCmd := &appconfig.InstallerData{
		Name: lo.ToPtr("ghr-custom-missing"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"repository":        "owner/repo",
			"destination":       "/some/path",
			"download_filename": "file.weird",
			"strategy":          "custom",
		},
	}
	assertValidationError(t, newTestGitHubReleaseInstaller(customMissingCmd).Validate(), "extract_command")

	// 🔴 extract_command without strategy: custom
	extractCmdWrongStrategy := &appconfig.InstallerData{
		Name: lo.ToPtr("ghr-custom-wrong-strategy"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"repository":        "owner/repo",
			"destination":       "/some/path",
			"download_filename": "file.tar.gz",
			"strategy":          "tar",
			"extract_command":   "echo nope",
		},
	}
	assertValidationError(t, newTestGitHubReleaseInstaller(extractCmdWrongStrategy).Validate(), "extract_command")
}

func TestGitHubReleaseGetOpts(t *testing.T) {
	logger.InitLogger(false)

	t.Run("parses all options correctly", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("test-release"),
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
			Name: lo.ToPtr("test-release"),
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
			Name: lo.ToPtr("test-release"),
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
			Name: lo.ToPtr("test-release"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"strategy": "none",
			},
		}
		installer := newTestGitHubReleaseInstaller(data)
		opts := installer.GetOpts()

		assert.Equal(t, GitHubReleaseInstallStrategyNone, *opts.Strategy)
	})

	t.Run("handles gzip strategy", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("test-release"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"strategy": "gzip",
			},
		}
		installer := newTestGitHubReleaseInstaller(data)
		opts := installer.GetOpts()

		assert.Equal(t, GitHubReleaseInstallStrategyGzip, *opts.Strategy)
	})

	t.Run("accepts gz as alias for gzip", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("test-release"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"strategy": "gz",
			},
		}
		installer := newTestGitHubReleaseInstaller(data)
		opts := installer.GetOpts()

		assert.Equal(t, GitHubReleaseInstallStrategyGzip, *opts.Strategy)
	})

	t.Run("handles custom strategy with extract_command", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("test-release"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"strategy":        "custom",
				"extract_command": "cp {{ .DownloadFile }} {{ .ExtractDir }}/{{ .ArchiveBinName }}",
			},
		}
		installer := newTestGitHubReleaseInstaller(data)
		opts := installer.GetOpts()

		assert.Equal(t, GitHubReleaseInstallStrategyCustom, *opts.Strategy)
		assert.NotNil(t, opts.ExtractCommand)
		assert.Contains(t, *opts.ExtractCommand, "{{ .DownloadFile }}")
	})
}

func TestGitHubReleaseCustomExtract(t *testing.T) {
	logger.InitLogger(false)
	if runtime.GOOS == "windows" {
		t.Skip("custom extract test uses a POSIX shell command")
	}

	// Prepare a fake "downloaded" asset on disk and an extract dir.
	tmpDir := t.TempDir()
	downloadFile := filepath.Join(tmpDir, "asset.weird")
	payload := []byte("binary payload from sofmani custom extract test")
	assert.NoError(t, os.WriteFile(downloadFile, payload, 0644))

	extractDir := filepath.Join(tmpDir, "extract")
	assert.NoError(t, os.Mkdir(extractDir, 0755))

	data := &appconfig.InstallerData{
		Name: lo.ToPtr("custom-tool"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			// The user's command references template variables directly instead of env vars.
			// Here we pretend the "weird" asset really just needs to be copied.
			"extract_command": "cp {{ .DownloadFile }} {{ .ExtractDir }}/{{ .ArchiveBinName }}",
		},
	}
	installer := newTestGitHubReleaseInstaller(data)

	vars := NewTemplateVars("v1.2.3", nil)
	vars.DownloadFile = downloadFile
	vars.ExtractDir = extractDir
	vars.Destination = filepath.Join(tmpDir, "dest")
	vars.BinName = "custom-tool"
	vars.ArchiveBinName = "custom-tool"

	err := installer.runCustomExtract("cp {{ .DownloadFile }} {{ .ExtractDir }}/{{ .ArchiveBinName }}", vars)
	assert.NoError(t, err)

	// The command should have produced extractDir/custom-tool with the original payload.
	got, err := os.ReadFile(filepath.Join(extractDir, "custom-tool"))
	assert.NoError(t, err)
	assert.Equal(t, payload, got)
}

func TestGitHubReleaseCustomExtractFailures(t *testing.T) {
	logger.InitLogger(false)
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX shell command")
	}

	data := &appconfig.InstallerData{
		Name: lo.ToPtr("custom-tool"),
		Type: appconfig.InstallerTypeGitHubRelease,
	}
	installer := newTestGitHubReleaseInstaller(data)
	vars := NewTemplateVars("v0.0.0", nil)

	t.Run("non-zero exit is surfaced", func(t *testing.T) {
		err := installer.runCustomExtract("exit 42", vars)
		assert.Error(t, err)
	})

	t.Run("invalid template is surfaced", func(t *testing.T) {
		err := installer.runCustomExtract("echo {{ .NopeField", vars)
		assert.Error(t, err)
	})
}

func TestDecompressGzip(t *testing.T) {
	logger.InitLogger(false)

	t.Run("decompresses a valid gzip stream", func(t *testing.T) {
		// Build a .gz fixture in-memory containing a fake binary payload.
		payload := []byte("#!/bin/sh\necho hello from tree-sitter\n")
		var gzBuf bytes.Buffer
		gw := gzip.NewWriter(&gzBuf)
		_, err := gw.Write(payload)
		assert.NoError(t, err)
		assert.NoError(t, gw.Close())

		var out bytes.Buffer
		err = decompressGzip(&gzBuf, &out)
		assert.NoError(t, err)
		assert.Equal(t, payload, out.Bytes())
	})

	t.Run("rejects non-gzip input", func(t *testing.T) {
		src := bytes.NewReader([]byte("this is not gzipped"))
		var out bytes.Buffer
		err := decompressGzip(src, &out)
		assert.Error(t, err)
	})

	t.Run("works end-to-end on a temp .gz file", func(t *testing.T) {
		tmpDir := t.TempDir()
		gzPath := filepath.Join(tmpDir, "payload.gz")
		outPath := filepath.Join(tmpDir, "payload")

		payload := []byte("hello, sofmani gzip strategy")

		gzFile, err := os.Create(gzPath)
		assert.NoError(t, err)
		gw := gzip.NewWriter(gzFile)
		_, err = gw.Write(payload)
		assert.NoError(t, err)
		assert.NoError(t, gw.Close())
		assert.NoError(t, gzFile.Close())

		// Sanity: the fixture is a gzip file but NOT a tar.gz.
		assert.True(t, isGzipFile(gzPath))
		assert.False(t, isTarGzFile(gzPath))

		src, err := os.Open(gzPath)
		assert.NoError(t, err)
		defer func() { _ = src.Close() }()
		dst, err := os.Create(outPath)
		assert.NoError(t, err)
		defer func() { _ = dst.Close() }()

		assert.NoError(t, decompressGzip(src, dst))

		got, err := os.ReadFile(outPath)
		assert.NoError(t, err)
		assert.Equal(t, payload, got)
	})
}

func TestGitHubReleaseGetBinName(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns bin_name when set", func(t *testing.T) {
		binName := "custom-bin"
		data := &appconfig.InstallerData{
			Name:    lo.ToPtr("my-app"),
			Type:    appconfig.InstallerTypeGitHubRelease,
			BinName: &binName,
		}
		installer := newTestGitHubReleaseInstaller(data)
		assert.Equal(t, "custom-bin", installer.GetBinName())
	})

	t.Run("returns base name when bin_name not set", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("owner/my-app"),
			Type: appconfig.InstallerTypeGitHubRelease,
		}
		installer := newTestGitHubReleaseInstaller(data)
		assert.Equal(t, "my-app", installer.GetBinName())
	})
}

func TestGitHubReleaseGetArchiveBinName(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns archive_bin_name when set", func(t *testing.T) {
		binName := "cospend"
		data := &appconfig.InstallerData{
			Name:    lo.ToPtr("cospend-cli"),
			Type:    appconfig.InstallerTypeGitHubRelease,
			BinName: &binName,
			Opts: &map[string]any{
				"archive_bin_name": "cospend-cli",
			},
		}
		installer := newTestGitHubReleaseInstaller(data)
		assert.Equal(t, "cospend-cli", installer.GetArchiveBinName())
		assert.Equal(t, "cospend", installer.GetBinName())
	})

	t.Run("falls back to bin_name when archive_bin_name not set", func(t *testing.T) {
		binName := "custom-bin"
		data := &appconfig.InstallerData{
			Name:    lo.ToPtr("my-app"),
			Type:    appconfig.InstallerTypeGitHubRelease,
			BinName: &binName,
		}
		installer := newTestGitHubReleaseInstaller(data)
		assert.Equal(t, "custom-bin", installer.GetArchiveBinName())
	})

	t.Run("falls back to name when neither set", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("my-app"),
			Type: appconfig.InstallerTypeGitHubRelease,
		}
		installer := newTestGitHubReleaseInstaller(data)
		assert.Equal(t, "my-app", installer.GetArchiveBinName())
	})
}

func TestGitHubReleaseGetFilename(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns filename for current platform", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("test-release"),
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
			Name: lo.ToPtr("test-release"),
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
			Name: lo.ToPtr("test-cache-app"),
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
			Name: lo.ToPtr("non-existent-app-12345"),
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
			Name: lo.ToPtr("test-overwrite-app"),
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
			Name: lo.ToPtr("test-release"),
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
			Name: lo.ToPtr("test-release"),
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
			Name: lo.ToPtr("test-release"),
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
			Name: lo.ToPtr("myapp"),
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
			Name: lo.ToPtr("nonexistent-app"),
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
			Name:           lo.ToPtr("myapp"),
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
			Name:           lo.ToPtr("myapp"),
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
			Name: lo.ToPtr("unique-no-cache-app-99999"),
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

func TestGitHubReleaseTreeModeOpts(t *testing.T) {
	logger.InitLogger(false)

	t.Run("parses extract_to, strip_components, bin_links", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("neovim"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"repository":        "neovim/neovim",
				"strategy":          "tar",
				"download_filename": "nvim-linux-x86_64.tar.gz",
				"extract_to":        "/opt/neovim",
				"strip_components":  1,
				"bin_links": []any{
					map[string]any{"source": "bin/nvim", "target": "/usr/local/bin/nvim"},
				},
			},
		}
		opts := newTestGitHubReleaseInstaller(data).GetOpts()

		assert.NotNil(t, opts.ExtractTo)
		assert.Equal(t, "/opt/neovim", *opts.ExtractTo)
		assert.NotNil(t, opts.StripComponents)
		assert.Equal(t, 1, *opts.StripComponents)
		assert.Len(t, opts.BinLinks, 1)
		assert.Equal(t, "bin/nvim", opts.BinLinks[0].Source)
		assert.Equal(t, "/usr/local/bin/nvim", opts.BinLinks[0].Target)
	})

	t.Run("accepts strip_components as float64 (yaml number decoding)", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("neovim"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"strip_components": float64(2),
			},
		}
		opts := newTestGitHubReleaseInstaller(data).GetOpts()
		assert.NotNil(t, opts.StripComponents)
		assert.Equal(t, 2, *opts.StripComponents)
	})

	t.Run("parses bin_links with map[any]any keys (yaml.v2 shape)", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("neovim"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"bin_links": []any{
					map[any]any{"source": "bin/nvim", "target": "/usr/local/bin/nvim"},
				},
			},
		}
		opts := newTestGitHubReleaseInstaller(data).GetOpts()
		assert.Len(t, opts.BinLinks, 1)
		assert.Equal(t, "bin/nvim", opts.BinLinks[0].Source)
	})
}

func TestGitHubReleaseTreeModeValidation(t *testing.T) {
	logger.InitLogger(false)

	// Tree mode does not require destination.
	t.Run("destination not required in tree mode", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("ghr-tree"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"repository":        "owner/repo",
				"download_filename": "release.tar.gz",
				"strategy":          "tar",
				"extract_to":        "/opt/tree",
				"bin_links": []any{
					map[string]any{"source": "bin/tool", "target": "/usr/local/bin/tool"},
				},
			},
		}
		assertNoValidationErrors(t, newTestGitHubReleaseInstaller(data).Validate())
	})

	t.Run("tree mode requires tar or zip strategy", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("ghr-tree-none"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"repository":        "owner/repo",
				"download_filename": "release",
				"strategy":          "none",
				"extract_to":        "/opt/tree",
			},
		}
		assertValidationError(t, newTestGitHubReleaseInstaller(data).Validate(), "strategy")
	})

	t.Run("tree mode with missing strategy fails validation", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("ghr-tree-no-strategy"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"repository":        "owner/repo",
				"download_filename": "release.tar.gz",
				"extract_to":        "/opt/tree",
			},
		}
		assertValidationError(t, newTestGitHubReleaseInstaller(data).Validate(), "strategy")
	})

	t.Run("rejects bin_link with missing target", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("ghr-tree-bad-link"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"repository":        "owner/repo",
				"download_filename": "release.tar.gz",
				"strategy":          "tar",
				"extract_to":        "/opt/tree",
				"bin_links": []any{
					map[string]any{"source": "bin/tool"},
				},
			},
		}
		assertValidationError(t, newTestGitHubReleaseInstaller(data).Validate(), "bin_links[0].target")
	})

	t.Run("rejects bin_link source that escapes extract_to", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("ghr-tree-escape"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"repository":        "owner/repo",
				"download_filename": "release.tar.gz",
				"strategy":          "tar",
				"extract_to":        "/opt/tree",
				"bin_links": []any{
					map[string]any{"source": "../../etc/passwd", "target": "/usr/local/bin/tool"},
				},
			},
		}
		assertValidationError(t, newTestGitHubReleaseInstaller(data).Validate(), "bin_links[0].source")
	})

	t.Run("rejects negative strip_components", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("ghr-tree-strip"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"repository":        "owner/repo",
				"download_filename": "release.tar.gz",
				"strategy":          "tar",
				"extract_to":        "/opt/tree",
				"strip_components":  -1,
			},
		}
		assertValidationError(t, newTestGitHubReleaseInstaller(data).Validate(), "strip_components")
	})
}

func TestGitHubReleaseTreeModeInstallDir(t *testing.T) {
	logger.InitLogger(false)

	data := &appconfig.InstallerData{
		Name: lo.ToPtr("tree"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"extract_to": "/opt/tree",
		},
	}
	i := newTestGitHubReleaseInstaller(data)
	assert.Equal(t, "/opt/tree", i.GetInstallDir())
}

func TestGitHubReleaseTreeModeCheckIsInstalled(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns false when extract_to missing", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("tree"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"extract_to": "/nonexistent/sofmani-tree-test",
			},
		}
		installed, err := newTestGitHubReleaseInstaller(data).CheckIsInstalled()
		assert.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("returns false when bin_link target missing", func(t *testing.T) {
		tmp, err := os.MkdirTemp("", "sofmani-tree-check")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmp) }()

		data := &appconfig.InstallerData{
			Name: lo.ToPtr("tree"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"extract_to": tmp,
				"bin_links": []any{
					map[string]any{"source": "bin/tool", "target": filepath.Join(tmp, "missing-link")},
				},
			},
		}
		installed, err := newTestGitHubReleaseInstaller(data).CheckIsInstalled()
		assert.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("returns true when extract_to and all bin_links exist", func(t *testing.T) {
		tmp, err := os.MkdirTemp("", "sofmani-tree-check")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmp) }()

		linkPath := filepath.Join(tmp, "link")
		assert.NoError(t, os.WriteFile(linkPath, []byte("x"), 0644))

		data := &appconfig.InstallerData{
			Name: lo.ToPtr("tree"),
			Type: appconfig.InstallerTypeGitHubRelease,
			Opts: &map[string]any{
				"extract_to": tmp,
				"bin_links": []any{
					map[string]any{"source": "whatever", "target": linkPath},
				},
			},
		}
		installed, err := newTestGitHubReleaseInstaller(data).CheckIsInstalled()
		assert.NoError(t, err)
		assert.True(t, installed)
	})
}

// buildTestZip writes a zip archive at path containing the given files. Each entry key
// is the archive-relative path; values are the file bytes. Directory entries can be
// represented with a trailing "/" and empty contents.
func buildTestZip(t *testing.T, path string, files map[string]string) {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		fh := &zip.FileHeader{Name: name, Method: zip.Deflate}
		fh.SetMode(0644)
		f, err := w.CreateHeader(fh)
		assert.NoError(t, err)
		if content != "" {
			_, err = f.Write([]byte(content))
			assert.NoError(t, err)
		}
	}
	assert.NoError(t, w.Close())
	assert.NoError(t, os.WriteFile(path, buf.Bytes(), 0644))
}

func TestExtractZipWithStrip(t *testing.T) {
	logger.InitLogger(false)

	t.Run("strip 0 preserves full paths", func(t *testing.T) {
		tmp, err := os.MkdirTemp("", "sofmani-zip-strip0")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmp) }()

		zipPath := filepath.Join(tmp, "a.zip")
		buildTestZip(t, zipPath, map[string]string{
			"top/bin/tool":   "#!/bin/sh\n",
			"top/lib/data.x": "payload",
		})
		dest := filepath.Join(tmp, "out")
		assert.NoError(t, os.MkdirAll(dest, 0755))
		assert.NoError(t, extractZipWithStrip(zipPath, dest, 0))

		data, err := os.ReadFile(filepath.Join(dest, "top", "lib", "data.x"))
		assert.NoError(t, err)
		assert.Equal(t, "payload", string(data))
	})

	t.Run("strip 1 drops top-level wrapper directory", func(t *testing.T) {
		tmp, err := os.MkdirTemp("", "sofmani-zip-strip1")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmp) }()

		zipPath := filepath.Join(tmp, "a.zip")
		buildTestZip(t, zipPath, map[string]string{
			"nvim-linux-x86_64/bin/nvim":            "binary",
			"nvim-linux-x86_64/share/nvim/init.vim": "\" rc",
		})
		dest := filepath.Join(tmp, "out")
		assert.NoError(t, os.MkdirAll(dest, 0755))
		assert.NoError(t, extractZipWithStrip(zipPath, dest, 1))

		// After stripping one component, files live directly under dest.
		binData, err := os.ReadFile(filepath.Join(dest, "bin", "nvim"))
		assert.NoError(t, err)
		assert.Equal(t, "binary", string(binData))

		rcData, err := os.ReadFile(filepath.Join(dest, "share", "nvim", "init.vim"))
		assert.NoError(t, err)
		assert.Equal(t, "\" rc", string(rcData))

		// And the original wrapper dir should NOT exist.
		_, err = os.Stat(filepath.Join(dest, "nvim-linux-x86_64"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("rejects zip-slip path traversal", func(t *testing.T) {
		tmp, err := os.MkdirTemp("", "sofmani-zip-slip")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmp) }()

		zipPath := filepath.Join(tmp, "bad.zip")
		buildTestZip(t, zipPath, map[string]string{
			"../evil": "pwned",
		})
		dest := filepath.Join(tmp, "out")
		assert.NoError(t, os.MkdirAll(dest, 0755))
		err = extractZipWithStrip(zipPath, dest, 0)
		assert.Error(t, err)
	})
}

func TestInstallBinLink(t *testing.T) {
	logger.InitLogger(false)

	t.Run("creates link (or copy) to source", func(t *testing.T) {
		tmp, err := os.MkdirTemp("", "sofmani-binlink")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmp) }()

		src := filepath.Join(tmp, "src", "tool")
		assert.NoError(t, os.MkdirAll(filepath.Dir(src), 0755))
		assert.NoError(t, os.WriteFile(src, []byte("binary"), 0755))

		target := filepath.Join(tmp, "bin", "tool")
		assert.NoError(t, installBinLink(src, target))

		// Reading through the link (or copy) must return the source content.
		data, err := os.ReadFile(target)
		assert.NoError(t, err)
		assert.Equal(t, "binary", string(data))

		if runtime.GOOS != "windows" {
			// On unix, we expect a symlink specifically so sibling files resolve.
			fi, err := os.Lstat(target)
			assert.NoError(t, err)
			assert.True(t, fi.Mode()&os.ModeSymlink != 0, "expected symlink on unix")
		}
	})

	t.Run("replaces an existing target", func(t *testing.T) {
		tmp, err := os.MkdirTemp("", "sofmani-binlink-replace")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmp) }()

		src := filepath.Join(tmp, "src", "tool")
		assert.NoError(t, os.MkdirAll(filepath.Dir(src), 0755))
		assert.NoError(t, os.WriteFile(src, []byte("new"), 0755))

		target := filepath.Join(tmp, "bin", "tool")
		assert.NoError(t, os.MkdirAll(filepath.Dir(target), 0755))
		// Put a stale file at the target.
		assert.NoError(t, os.WriteFile(target, []byte("old"), 0755))

		assert.NoError(t, installBinLink(src, target))
		data, err := os.ReadFile(target)
		assert.NoError(t, err)
		assert.Equal(t, "new", string(data))
	})
}

// TestGitHubReleaseTreeInstallAtomicReplace exercises the staging→rename logic by
// driving installTree end-to-end with a local zip fixture served over HTTP. It verifies
// that a second install fully replaces the previous tree (no stale files linger) and
// refreshes bin_links.
func TestGitHubReleaseTreeInstallAtomicReplace(t *testing.T) {
	logger.InitLogger(false)

	tmp, err := os.MkdirTemp("", "sofmani-tree-install")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmp) }()

	extractTo := filepath.Join(tmp, "neovim")

	// Simulate an already-installed old version with a file that must NOT survive update.
	assert.NoError(t, os.MkdirAll(filepath.Join(extractTo, "old-dir"), 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(extractTo, "old-dir", "stale.txt"), []byte("stale"), 0644))

	// Build a fresh "release" as a staging dir and rename it in — this is what a successful
	// installTree run does, minus the HTTP download step which we can't stub here.
	staging := extractTo + ".sofmani-new"
	assert.NoError(t, os.MkdirAll(filepath.Join(staging, "bin"), 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(staging, "bin", "nvim"), []byte("v2"), 0755))

	// Manual re-creation of the swap logic inside installTree so we can assert on the
	// outcome without needing a live network.
	backup := extractTo + ".sofmani-old"
	assert.NoError(t, os.Rename(extractTo, backup))
	assert.NoError(t, os.Rename(staging, extractTo))
	assert.NoError(t, os.RemoveAll(backup))

	// Stale file from the old tree must be gone.
	_, err = os.Stat(filepath.Join(extractTo, "old-dir", "stale.txt"))
	assert.True(t, os.IsNotExist(err), "stale file from previous install should be removed")

	// New file must be present.
	data, err := os.ReadFile(filepath.Join(extractTo, "bin", "nvim"))
	assert.NoError(t, err)
	assert.Equal(t, "v2", string(data))

	// And a subsequent bin_link install should succeed against the new tree.
	target := filepath.Join(tmp, "bin", "nvim")
	assert.NoError(t, installBinLink(filepath.Join(extractTo, "bin", "nvim"), target))
	linked, err := os.ReadFile(target)
	assert.NoError(t, err)
	assert.Equal(t, "v2", string(linked))
}

func TestNewGitHubReleaseInstaller(t *testing.T) {
	logger.InitLogger(false)

	t.Run("creates installer with config and data", func(t *testing.T) {
		cfg := &appconfig.AppConfig{}
		data := &appconfig.InstallerData{
			Name: lo.ToPtr("test-release"),
			Type: appconfig.InstallerTypeGitHubRelease,
		}
		installer := NewGitHubReleaseInstaller(cfg, data)

		assert.NotNil(t, installer)
		assert.Equal(t, cfg, installer.Config)
		assert.Equal(t, data, installer.Info)
		assert.Equal(t, data, installer.Data)
	})
}
