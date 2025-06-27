package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRealPath(t *testing.T) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	env := []string{"GOPATH=/go"} // HOME is now handled by the function itself or by os.UserHomeDir()
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"Expand home directory", "~/project", filepath.Join(userHomeDir, "project")},
		{"Expand home directory with trailing slash", "~/project/", filepath.Join(userHomeDir, "project") + string(filepath.Separator)},
		{"Expand home directory only", "~", userHomeDir},
		{"Expand GOPATH", "$GOPATH/src", "/go/src"},
		{"No expansion", "/usr/local/bin", "/usr/local/bin"},
		{"Path with spaces", "/my path/with spaces", "/my path/with spaces"},
		{"Path with environment variable and spaces", "$GOPATH/my src", "/go/my src"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentEnv := env
			result := GetRealPath(currentEnv, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
