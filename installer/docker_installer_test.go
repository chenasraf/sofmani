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

func TestDockerVersionPin(t *testing.T) {
	logger.InitLogger(false)

	unpinned := newTestDockerInstaller(&appconfig.InstallerData{
		Name: lo.ToPtr("ghcr.io/open-webui/open-webui"),
		Type: appconfig.InstallerTypeDocker,
	})
	require.Equal(t, "", unpinned.GetPinnedVersion())
	require.Equal(t, "ghcr.io/open-webui/open-webui", unpinned.GetImage())

	pinned := newTestDockerInstaller(&appconfig.InstallerData{
		Name: lo.ToPtr("ghcr.io/open-webui/open-webui"),
		Type: appconfig.InstallerTypeDocker,
		Opts: &map[string]any{"version": "v0.5.0"},
	})
	require.Equal(t, "v0.5.0", pinned.GetPinnedVersion())
	require.Equal(t, "ghcr.io/open-webui/open-webui:v0.5.0", pinned.GetImage())

	// A tag on the image name may be a moving one, so it keeps following the registry.
	tagged := newTestDockerInstaller(&appconfig.InstallerData{
		Name: lo.ToPtr("ghcr.io/open-webui/open-webui:main"),
		Type: appconfig.InstallerTypeDocker,
		Opts: &map[string]any{"version": "v0.5.0"},
	})
	require.Equal(t, "", tagged.GetPinnedVersion())
	require.Equal(t, "ghcr.io/open-webui/open-webui:main", tagged.GetImage())
}

func TestImageHasTag(t *testing.T) {
	require.False(t, imageHasTag("nginx"))
	require.True(t, imageHasTag("nginx:1.25"))
	require.False(t, imageHasTag("ghcr.io/owner/image"))
	require.True(t, imageHasTag("ghcr.io/owner/image:latest"))
	// A registry port is not a tag.
	require.False(t, imageHasTag("localhost:5000/owner/image"))
	require.True(t, imageHasTag("localhost:5000/owner/image:1.0"))
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
