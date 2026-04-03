package installer

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/chenasraf/sofmani/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrequencyCacheFileName(t *testing.T) {
	assert.Equal(t, "freq_my-installer", frequencyCacheFileName("my-installer"))
	assert.Equal(t, "freq_path__to__thing", frequencyCacheFileName("path/to/thing"))
	assert.Equal(t, "freq_has_spaces", frequencyCacheFileName("has spaces"))
	assert.Equal(t, "freq_back__slash", frequencyCacheFileName("back\\slash"))
}

func TestCheckFrequency_NoPreviousRun(t *testing.T) {
	shouldRun, err := checkFrequency("nonexistent-installer-test", "1d")
	assert.NoError(t, err)
	assert.True(t, shouldRun)
}

func TestCheckFrequency_RecentRun(t *testing.T) {
	name := "test-freq-recent"
	cacheDir, err := utils.GetCacheDir()
	require.NoError(t, err)

	cacheFile := filepath.Join(cacheDir, frequencyCacheFileName(name))
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	err = os.WriteFile(cacheFile, []byte(ts), 0644)
	require.NoError(t, err)
	defer func() { _ = os.Remove(cacheFile) }()

	shouldRun, err := checkFrequency(name, "1d")
	assert.NoError(t, err)
	assert.False(t, shouldRun)
}

func TestCheckFrequency_ExpiredRun(t *testing.T) {
	name := "test-freq-expired"
	cacheDir, err := utils.GetCacheDir()
	require.NoError(t, err)

	cacheFile := filepath.Join(cacheDir, frequencyCacheFileName(name))
	ts := strconv.FormatInt(time.Now().Add(-48*time.Hour).Unix(), 10)
	err = os.WriteFile(cacheFile, []byte(ts), 0644)
	require.NoError(t, err)
	defer func() { _ = os.Remove(cacheFile) }()

	shouldRun, err := checkFrequency(name, "1d")
	assert.NoError(t, err)
	assert.True(t, shouldRun)
}

func TestWriteFrequencyTimestamp(t *testing.T) {
	name := "test-freq-write"
	cacheDir, err := utils.GetCacheDir()
	require.NoError(t, err)

	cacheFile := filepath.Join(cacheDir, frequencyCacheFileName(name))
	defer func() { _ = os.Remove(cacheFile) }()

	err = writeFrequencyTimestamp(name)
	assert.NoError(t, err)

	data, err := os.ReadFile(cacheFile)
	require.NoError(t, err)

	ts, err := strconv.ParseInt(string(data), 10, 64)
	require.NoError(t, err)

	assert.InDelta(t, time.Now().Unix(), ts, 2)
}
