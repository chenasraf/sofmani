package utils

import "strings"

func IsGitURL(url string) bool {
	return strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "git@")
}
