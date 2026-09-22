package installer

import (
	"os"
	"testing"

	"github.com/chenasraf/sofmani/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheFileName(t *testing.T) {
	assert.Equal(t, "version_my-installer", cacheFileName("version_", "my-installer"))
	assert.Equal(t, "version_path__to__thing", cacheFileName("version_", "path/to/thing"))
	assert.Equal(t, "version_ghcr.io__owner__image", cacheFileName("version_", "ghcr.io/owner:image"))
	assert.Equal(t, "freq_has_spaces", cacheFileName("freq_", "has spaces"))
}

// recordTestVersion writes a version record and removes it when the test ends.
func recordTestVersion(t *testing.T, name string, version string) {
	t.Helper()
	require.NoError(t, RecordInstalledVersion(name, version))
	t.Cleanup(func() {
		file, err := pinnedVersionCacheFile(name)
		if err == nil {
			_ = os.Remove(file)
		}
	})
}

func TestReadInstalledVersion(t *testing.T) {
	logger.InitLogger(false)

	assert.Empty(t, ReadInstalledVersion("test-pin-never-installed"))

	recordTestVersion(t, "test-pin-read", "1.2.3")
	assert.Equal(t, "1.2.3", ReadInstalledVersion("test-pin-read"))
}

func TestClearInstalledVersion(t *testing.T) {
	logger.InitLogger(false)

	recordTestVersion(t, "test-pin-clear", "1.2.3")
	ClearInstalledVersion("test-pin-clear")
	assert.Empty(t, ReadInstalledVersion("test-pin-clear"))

	// Clearing what was never recorded is not an error.
	ClearInstalledVersion("test-pin-clear-missing")
}

func TestPinnedVersionNeedsUpdate(t *testing.T) {
	logger.InitLogger(false)

	// Nothing recorded: the package is reconciled to the pin.
	assert.True(t, PinnedVersionNeedsUpdate("test-pin-unrecorded", "1.2.3"))

	recordTestVersion(t, "test-pin-compare", "1.2.3")
	assert.False(t, PinnedVersionNeedsUpdate("test-pin-compare", "1.2.3"))
	assert.True(t, PinnedVersionNeedsUpdate("test-pin-compare", "1.3.0"))
}
