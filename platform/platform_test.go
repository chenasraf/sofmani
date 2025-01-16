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

func TestGetShouldRunOnOSOnly(t *testing.T) {
	platforms := Platforms{Only: &[]Platform{PlatformMacos}}
	assert.True(t, platforms.GetShouldRunOnOS(PlatformMacos))
	assert.False(t, platforms.GetShouldRunOnOS(PlatformLinux))
}

func TestGetShouldRunOnOSExcept(t *testing.T) {
	platforms := Platforms{Except: &[]Platform{PlatformMacos}}
	assert.False(t, platforms.GetShouldRunOnOS(PlatformMacos))
	assert.True(t, platforms.GetShouldRunOnOS(PlatformLinux))
}
