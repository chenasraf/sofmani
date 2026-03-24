package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func newTestDockerInstaller(data *appconfig.InstallerData) *DockerInstaller {
	return &DockerInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config: nil,
		Info:   data,
	}
}

func TestDockerValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid: just name and type
	validData := &appconfig.InstallerData{
		Name:    lo.ToPtr("ghcr.io/open-webui/open-webui:main"),
		Type:    appconfig.InstallerTypeDocker,
		BinName: lo.ToPtr("open-webui"),
	}
	assertNoValidationErrors(t, newTestDockerInstaller(validData).Validate())

	// 🟢 Valid: with flags
	withFlags := &appconfig.InstallerData{
		Name:    lo.ToPtr("ghcr.io/open-webui/open-webui:main"),
		Type:    appconfig.InstallerTypeDocker,
		BinName: lo.ToPtr("open-webui"),
		Opts: &map[string]any{
			"flags": "-p 3300:8080 -v open-webui:/data",
		},
	}
	assertNoValidationErrors(t, newTestDockerInstaller(withFlags).Validate())

	// 🔴 Invalid: missing name (should be caught by BaseValidate)
	invalid := &appconfig.InstallerData{
		Type: appconfig.InstallerTypeDocker,
	}
	assertValidationError(t, newTestDockerInstaller(invalid).Validate(), "name")
}

func TestExtractDigestFromManifest(t *testing.T) {
	data := []byte(`{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
		"manifests": [
			{
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"digest": "sha256:abc",
				"platform": {
					"architecture": "arm64",
					"os": "darwin"
				}
			},
			{
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"digest": "sha256:def",
				"platform": {
					"architecture": "amd64",
					"os": "linux"
				}
			}
		]
	}`)

	digest, err := extractDigestFromManifest(data, "darwin", "arm64")
	require.NoError(t, err)
	require.Equal(t, "abc", digest)

	digest, err = extractDigestFromManifest(data, "linux", "amd64")
	require.NoError(t, err)
	require.Equal(t, "def", digest)
}
