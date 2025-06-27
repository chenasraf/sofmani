package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
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
}
