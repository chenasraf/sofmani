package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCacheDir(t *testing.T) {
	t.Run("returns a valid path", func(t *testing.T) {
		cacheDir, err := GetCacheDir()
		assert.NoError(t, err)
		assert.NotEmpty(t, cacheDir)
	})

	t.Run("path ends with sofmani", func(t *testing.T) {
		cacheDir, err := GetCacheDir()
		assert.NoError(t, err)
		assert.True(t, strings.HasSuffix(cacheDir, "sofmani"))
	})

	t.Run("creates the directory if it does not exist", func(t *testing.T) {
		cacheDir, err := GetCacheDir()
		assert.NoError(t, err)

		// Check directory exists
		info, err := os.Stat(cacheDir)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("directory is in user cache directory", func(t *testing.T) {
		cacheDir, err := GetCacheDir()
		assert.NoError(t, err)

		userCacheDir, err := os.UserCacheDir()
		assert.NoError(t, err)

		expectedPath := filepath.Join(userCacheDir, "sofmani")
		assert.Equal(t, expectedPath, cacheDir)
	})

	t.Run("returns same path on subsequent calls", func(t *testing.T) {
		cacheDir1, err := GetCacheDir()
		assert.NoError(t, err)

		cacheDir2, err := GetCacheDir()
		assert.NoError(t, err)

		assert.Equal(t, cacheDir1, cacheDir2)
	})
}
