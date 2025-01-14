package platform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentPlatform(t *testing.T) {
	pl := GetPlatform()
	assert.Contains(t, []Platform{PlatformMacos, PlatformLinux, PlatformWindows}, pl)
}

func TestContainsPlatform(t *testing.T) {
	platforms := []Platform{PlatformMacos, PlatformLinux}
	assert.True(t, ContainsPlatform(&platforms, PlatformMacos))
	assert.False(t, ContainsPlatform(&platforms, PlatformWindows))
}
