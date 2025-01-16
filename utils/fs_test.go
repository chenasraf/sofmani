package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRealPath(t *testing.T) {
	env := []string{"HOME=/home/user", "GOPATH=/go"}
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"Expand home directory", "~/project", "/home/user/project"},
		{"Expand GOPATH", "$GOPATH/src", "/go/src"},
		{"No expansion", "/usr/local/bin", "/usr/local/bin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRealPath(env, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
