package utils

import (
	"os"
	"path/filepath"
)

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
