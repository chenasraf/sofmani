package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// GetRealPath resolves a path string, expanding environment variables and replacing "~" with the user's home directory.
// It takes a slice of environment variables to temporarily set during expansion.
func GetRealPath(env []string, path string) string {
	// Temporarily set environment variables for expansion
	originalEnv := map[string]string{}
	for _, e := range env {
		split := strings.SplitN(e, "=", 2) // Use SplitN to handle cases where value might contain "="
		if len(split) == 2 {
			k, v := split[0], split[1]
			originalEnv[k] = os.Getenv(k) // Store original value to restore later
			os.Setenv(k, v)
		}
	}

	path = os.ExpandEnv(path) // Expand environment variables like $HOME or %USERPROFILE%

	// Restore original environment variables
	for k, v := range originalEnv {
		if v == "" {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, v)
		}
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, fmt.Sprintf("~%s", string(filepath.Separator))) || path == "~" {
		homedir, err := os.UserHomeDir()
		if err == nil { // Only replace if UserHomeDir succeeds
			if path == "~" {
				path = homedir
			} else {
				isDir := strings.HasSuffix(path, string(filepath.Separator))
				path = filepath.Join(homedir, path[2:])
				if isDir && !strings.HasSuffix(path, string(filepath.Separator)) { // Ensure trailing slash is preserved if originally present
					path += string(filepath.Separator)
				}
			}
		}
	}
	return strings.TrimSpace(path)
}

// PathExists checks if a file or directory exists at the given path.
// It returns true if the path exists, false otherwise.
// It does not distinguish between files and directories.
// An error is returned if os.Stat encounters an issue other than fs.ErrNotExist.
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil // Path exists
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil // Path does not exist, no error
	}
	return false, err // Other error (e.g., permission denied)
}
