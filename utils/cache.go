package utils

import (
	"os"
	"path/filepath"
)

// GetCacheDir returns the path to the user's cache directory for the application.
// It creates the directory if it doesn't exist.
// The cache directory is typically located at `~/.config/.cache` on Linux/macOS
// and `%APPDATA%\.cache` on Windows.
func GetCacheDir() (string, error) {
	confDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	cacheDir := filepath.Join(confDir, ".cache")
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return "", err
	}
	return cacheDir, nil
}
