package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
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
