package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsGitURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"Valid HTTPS URL", "https://github.com/user/repo.git", true},
		{"Valid SSH URL", "git@github.com:user/repo.git", true},
		{"URL ending with .git", "ftp://example.com/user/repo.git", true},
		{"Empty URL", "", false},
		{"Random string", "not_a_url", false},
		{"GitHub URL without .git", "https://github.com/user/repo", true},
		{"GitLab URL without .git", "https://gitlab.com/user/repo", true},
		{"Raw GitHub URL", "https://raw.githubusercontent.com/user/repo/main/file.txt", false},
		{"Raw GitLab URL", "https://gitlab.com/user/repo/-/raw/main/file.txt", false},
		{"Generic HTTPS URL", "https://example.com/file.yaml", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGitURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseGitURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
		expected    *GitURLInfo
	}{
		{
			name: "GitHub SSH",
			url:  "git@github.com:chenasraf/sofmani.git",
			expected: &GitURLInfo{
				Host:  "github.com",
				Owner: "chenasraf",
				Repo:  "sofmani",
			},
		},
		{
			name: "GitHub HTTPS with .git",
			url:  "https://github.com/chenasraf/sofmani.git",
			expected: &GitURLInfo{
				Host:  "github.com",
				Owner: "chenasraf",
				Repo:  "sofmani",
			},
		},
		{
			name: "GitHub HTTPS without .git",
			url:  "https://github.com/chenasraf/sofmani",
			expected: &GitURLInfo{
				Host:  "github.com",
				Owner: "chenasraf",
				Repo:  "sofmani",
			},
		},
		{
			name: "GitLab SSH",
			url:  "git@gitlab.com:user/project.git",
			expected: &GitURLInfo{
				Host:  "gitlab.com",
				Owner: "user",
				Repo:  "project",
			},
		},
		{
			name:        "Invalid URL",
			url:         "not-a-valid-url",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseGitURL(tt.url)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetRawFileURL(t *testing.T) {
	tests := []struct {
		name        string
		gitURL      string
		ref         string
		path        string
		expectError bool
		expected    string
	}{
		{
			name:     "GitHub SSH with ref",
			gitURL:   "git@github.com:chenasraf/sofmani.git",
			ref:      "master",
			path:     "docs/recipes/lazygit.yml",
			expected: "https://raw.githubusercontent.com/chenasraf/sofmani/master/docs/recipes/lazygit.yml",
		},
		{
			name:     "GitHub HTTPS default ref",
			gitURL:   "https://github.com/chenasraf/sofmani.git",
			ref:      "",
			path:     "README.md",
			expected: "https://raw.githubusercontent.com/chenasraf/sofmani/master/README.md",
		},
		{
			name:     "GitLab SSH",
			gitURL:   "git@gitlab.com:user/project.git",
			ref:      "develop",
			path:     "config.yml",
			expected: "https://gitlab.com/user/project/-/raw/develop/config.yml",
		},
		{
			name:     "Bitbucket HTTPS",
			gitURL:   "https://bitbucket.org/user/repo.git",
			ref:      "main",
			path:     "file.txt",
			expected: "https://bitbucket.org/user/repo/raw/main/file.txt",
		},
		{
			name:     "Path with leading slash",
			gitURL:   "git@github.com:user/repo.git",
			ref:      "main",
			path:     "/path/to/file.yml",
			expected: "https://raw.githubusercontent.com/user/repo/main/path/to/file.yml",
		},
		{
			name:     "Self-hosted GitLab",
			gitURL:   "git@gitlab.company.com:team/project.git",
			ref:      "main",
			path:     "manifest.yml",
			expected: "https://gitlab.company.com/team/project/-/raw/main/manifest.yml",
		},
		{
			name:     "Self-hosted Bitbucket",
			gitURL:   "https://bitbucket.mycompany.org/user/repo.git",
			ref:      "master",
			path:     "config.yml",
			expected: "https://bitbucket.mycompany.org/user/repo/raw/master/config.yml",
		},
		{
			name:     "Unknown host falls back to GitLab style",
			gitURL:   "git@custom.host.com:user/repo.git",
			ref:      "main",
			path:     "file.txt",
			expected: "https://custom.host.com/user/repo/-/raw/main/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetRawFileURL(tt.gitURL, tt.ref, tt.path)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestIsRawFileURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"GitHub raw URL", "https://raw.githubusercontent.com/user/repo/main/file.txt", true},
		{"GitLab raw URL", "https://gitlab.com/user/repo/-/raw/main/file.txt", true},
		{"Bitbucket raw URL", "https://bitbucket.org/user/repo/raw/main/file.txt", true},
		{"Regular GitHub URL", "https://github.com/user/repo", false},
		{"Regular GitLab URL", "https://gitlab.com/user/repo", false},
		{"Local path", "/path/to/file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRawFileURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectGitHostType(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected GitHostType
	}{
		{"GitHub", "github.com", GitHostGitHub},
		{"GitHub Enterprise", "github.mycompany.com", GitHostGitHub},
		{"GitLab", "gitlab.com", GitHostGitLab},
		{"Self-hosted GitLab", "gitlab.company.org", GitHostGitLab},
		{"Bitbucket", "bitbucket.org", GitHostBitbucket},
		{"Self-hosted Bitbucket", "bitbucket.mycompany.com", GitHostBitbucket},
		{"Unknown host", "git.custom.com", GitHostUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectGitHostType(tt.host)
			assert.Equal(t, tt.expected, result)
		})
	}
}
