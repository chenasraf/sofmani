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
		{"Invalid URL", "ftp://github.com/user/repo.git", false},
		{"Empty URL", "", false},
		{"Random string", "not_a_url", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGitURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}
