package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// IsGitURL checks if a string is likely a Git repository URL (not a raw file URL).
// It checks for "git@" prefix or URLs ending with ".git".
func IsGitURL(url string) bool {
	if strings.HasPrefix(url, "git@") {
		return true
	}
	if strings.HasSuffix(url, ".git") {
		return true
	}
	// Check for common git hosting patterns (but not raw file URLs)
	gitPatterns := []string{
		"github.com/",
		"gitlab.com/",
		"bitbucket.org/",
	}
	for _, pattern := range gitPatterns {
		if strings.Contains(url, pattern) && !IsRawFileURL(url) {
			return true
		}
	}
	return false
}

// IsRawFileURL checks if a URL is a direct raw file URL.
func IsRawFileURL(url string) bool {
	rawPatterns := []string{
		"raw.githubusercontent.com",
		"/-/raw/", // GitLab raw URL pattern
		"/raw/",   // Bitbucket raw URL pattern
		"raw.github.com",
	}
	for _, pattern := range rawPatterns {
		if strings.Contains(url, pattern) {
			return true
		}
	}
	return false
}

// GitURLInfo holds parsed information from a Git URL.
type GitURLInfo struct {
	Host  string // e.g., "github.com", "gitlab.com"
	Owner string // e.g., "chenasraf"
	Repo  string // e.g., "sofmani"
}

// ParseGitURL parses a Git URL (SSH or HTTPS) and extracts host, owner, and repo.
// Supports formats:
//   - git@github.com:owner/repo.git
//   - https://github.com/owner/repo.git
//   - https://github.com/owner/repo
func ParseGitURL(url string) (*GitURLInfo, error) {
	// SSH format: git@host:owner/repo.git
	sshRegex := regexp.MustCompile(`^git@([^:]+):([^/]+)/(.+?)(?:\.git)?$`)
	if matches := sshRegex.FindStringSubmatch(url); matches != nil {
		return &GitURLInfo{
			Host:  matches[1],
			Owner: matches[2],
			Repo:  matches[3],
		}, nil
	}

	// HTTPS format: https://host/owner/repo.git or https://host/owner/repo
	httpsRegex := regexp.MustCompile(`^https://([^/]+)/([^/]+)/(.+?)(?:\.git)?$`)
	if matches := httpsRegex.FindStringSubmatch(url); matches != nil {
		return &GitURLInfo{
			Host:  matches[1],
			Owner: matches[2],
			Repo:  matches[3],
		}, nil
	}

	return nil, fmt.Errorf("unable to parse Git URL: %s", url)
}

// GitHostType represents the type of Git hosting service.
type GitHostType string

const (
	GitHostGitHub    GitHostType = "github"
	GitHostGitLab    GitHostType = "gitlab"
	GitHostBitbucket GitHostType = "bitbucket"
	GitHostUnknown   GitHostType = "unknown"
)

// DetectGitHostType detects the Git hosting service from a host string.
// It checks for known patterns in the hostname.
func DetectGitHostType(host string) GitHostType {
	host = strings.ToLower(host)
	switch {
	case strings.Contains(host, "github"):
		return GitHostGitHub
	case strings.Contains(host, "gitlab"):
		return GitHostGitLab
	case strings.Contains(host, "bitbucket"):
		return GitHostBitbucket
	default:
		return GitHostUnknown
	}
}

// GetRawFileURL converts a Git repository URL to a raw file URL for direct access.
// Supports GitHub, GitLab (including self-hosted), and Bitbucket (including self-hosted).
// For unknown hosts, it attempts to use GitLab-style raw URLs as a fallback.
// Parameters:
//   - gitURL: The Git repository URL (SSH or HTTPS)
//   - ref: The branch, tag, or commit (defaults to "master" if empty)
//   - path: The file path within the repository
func GetRawFileURL(gitURL, ref, path string) (string, error) {
	info, err := ParseGitURL(gitURL)
	if err != nil {
		return "", err
	}

	if ref == "" {
		ref = "master"
	}

	// Remove leading slash from path if present
	path = strings.TrimPrefix(path, "/")

	hostType := DetectGitHostType(info.Host)

	switch hostType {
	case GitHostGitHub:
		// GitHub: https://raw.githubusercontent.com/owner/repo/ref/path
		return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", info.Owner, info.Repo, ref, path), nil
	case GitHostGitLab:
		// GitLab (including self-hosted): https://host/owner/repo/-/raw/ref/path
		return fmt.Sprintf("https://%s/%s/%s/-/raw/%s/%s", info.Host, info.Owner, info.Repo, ref, path), nil
	case GitHostBitbucket:
		// Bitbucket (including self-hosted): https://host/owner/repo/raw/ref/path
		return fmt.Sprintf("https://%s/%s/%s/raw/%s/%s", info.Host, info.Owner, info.Repo, ref, path), nil
	default:
		// For unknown hosts, try GitLab-style as it's common for self-hosted instances
		return fmt.Sprintf("https://%s/%s/%s/-/raw/%s/%s", info.Host, info.Owner, info.Repo, ref, path), nil
	}
}
