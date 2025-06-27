package utils

import "strings"

// IsGitURL checks if a string is likely a Git URL.
// It checks for "https://" or "git@" prefixes.
func IsGitURL(url string) bool {
	return strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "git@")
}
