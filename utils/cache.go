package utils

import (
	"os"
	"path/filepath"
)

// GetCacheDir returns the path to the user's cache directory for the application.
// It creates the directory if it doesn't exist.
// The cache directory is typically located at `$XDG_CACHE_HOME` on Linux/macOS
// and `%LocalAppData%` on Windows.
func GetCacheDir() (string, error) {
	confDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	cacheDir := filepath.Join(confDir, "sofmani")
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return "", err
	}
	return cacheDir, nil
}
