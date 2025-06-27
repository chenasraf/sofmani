package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
)

func newTestGitInstaller(data *appconfig.InstallerData) *GitInstaller {
	return &GitInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Info: data,
	}
}

func TestGitValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid: Both destination and ref are present
	validData := &appconfig.InstallerData{
		Name: strPtr("test-git-valid"),
		Type: appconfig.InstallerTypeGit,
		Opts: &map[string]any{
			"destination": "/some/path",
			"ref":         "main",
		},
	}
	assertNoValidationErrors(t, newTestGitInstaller(validData).Validate())

	// 🟢 Valid: Missing ref
	missingRefData := &appconfig.InstallerData{
		Name: strPtr("test-git-missing-ref"),
		Type: appconfig.InstallerTypeGit,
		Opts: &map[string]any{
			"destination": "/some/path",
		},
	}
	assertNoValidationErrors(t, newTestGitInstaller(missingRefData).Validate())

	// 🔴 Invalid: Missing destination
	missingDestData := &appconfig.InstallerData{
		Name: strPtr("test-git-missing-destination"),
		Type: appconfig.InstallerTypeGit,
		Opts: &map[string]any{
			"ref": "main",
		},
	}
	assertValidationError(t, newTestGitInstaller(missingDestData).Validate(), "destination")
}
