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
	return filepath.Join(confDir, ".cache"), nil
}
