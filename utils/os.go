package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetRealPath(path string) string {
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
	return path
}
