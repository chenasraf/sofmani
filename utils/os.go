package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func GetRealPath(env []string, path string) string {
	for _, e := range env {
		split := strings.Split(e, "=")
		k, v := split[0], split[1]
		os.Setenv(k, v)
	}
	path = os.ExpandEnv(path)
	if strings.HasPrefix(path, fmt.Sprintf("~%s", string(filepath.Separator))) {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		isDir := false
		if strings.HasSuffix(path, string(filepath.Separator)) {
			isDir = true
		}
		path = filepath.Join(homedir, path[2:])
		if isDir {
			path += string(filepath.Separator)
		}
	}
	return strings.TrimSpace(string(path))
}

func PathExists(path string) (error, bool) {
	_, err := os.Stat(path)
	exists := !errors.Is(err, fs.ErrNotExist)
	return nil, exists
}
