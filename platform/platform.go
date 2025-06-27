package platform

import (
	"fmt"
	"runtime"
	"slices"
)

var osValue string = runtime.GOOS // osValue stores the current operating system.

// getOS returns the current operating system. It caches the value after the first call.
func getOS() string {
	if osValue == "" {
		osValue = runtime.GOOS
	}
	return osValue
}

// SetOS overrides the detected operating system. This is primarily used for testing.
func SetOS(v string) {
	osValue = v
}

// GetPlatform returns the current platform (macos, linux, or windows).
func GetPlatform() Platform {
	switch getOS() {
	case "darwin":
		return PlatformMacos
	case "linux":
		return PlatformLinux
	case "windows":
		return PlatformWindows
	}
	panic(fmt.Sprintf("Unsupported platform %s", getOS()))
}

// Platforms defines which platforms a configuration applies to.
type Platforms struct {
	// Only specifies a list of platforms where the configuration should apply.
	Only *[]Platform `json:"only"   yaml:"only"`
	// Except specifies a list of platforms where the configuration should not apply.
	Except *[]Platform `json:"except" yaml:"except"`
}

// Platform represents an operating system platform.
type Platform string

// Constants for supported platforms.
const (
	PlatformMacos   Platform = "macos"   // PlatformMacos represents macOS.
	PlatformLinux   Platform = "linux"   // PlatformLinux represents Linux.
	PlatformWindows Platform = "windows" // PlatformWindows represents Windows.
)

// PlatformMap is a generic type that holds platform-specific values.
type PlatformMap[T any] struct {
	// MacOS is the value for macOS.
	MacOS *T `json:"macos"   yaml:"macos"`
	// Linux is the value for Linux.
	Linux *T `json:"linux"   yaml:"linux"`
	// Windows is the value for Windows.
	Windows *T `json:"windows" yaml:"windows"`
}

// Resolve returns the value for the current platform from the PlatformMap.
// It returns nil if no value is defined for the current platform.
func (p *PlatformMap[T]) Resolve() *T {
	if p == nil {
		return nil
	}
	switch getOS() {
	case "darwin":
		if p.MacOS != nil {
			return p.MacOS
		}
		return nil
	case "linux":
		if p.Linux != nil {
			return p.Linux
		}
		return nil
	case "windows":
		if p.Windows != nil {
			return p.Windows
		}
		return nil
	default:
		return nil
	}
}

// ResolveWithFallback returns the value for the current platform from the PlatformMap.
// If no value is defined for the current platform, it falls back to the value from the provided fallback PlatformMap.
func (o *PlatformMap[T]) ResolveWithFallback(fallback PlatformMap[T]) T {
	val := o.Resolve()
	if val == nil {
		return *fallback.Resolve()
	}
	return *val
}

// ContainsPlatform checks if a slice of platforms contains a specific platform.
func ContainsPlatform(platforms *[]Platform, platform Platform) bool {
	return slices.Contains(*platforms, platform)
}

// GetShouldRunOnOS determines if a configuration should run on the current operating system
// based on the Only and Except fields of the Platforms struct.
func (p *Platforms) GetShouldRunOnOS(curOS Platform) bool {
	if p == nil {
		return true
	}

	if p.Only != nil {
		return ContainsPlatform(p.Only, curOS)
	}
	if p.Except != nil {
		return !ContainsPlatform(p.Except, curOS)
	}
	return true
}

// strPtr returns a pointer to a string.
func strPtr(s string) *string {
	return &s
}

// DockerOSMap is a PlatformMap that defines the Docker OS for each platform.
var DockerOSMap = PlatformMap[string]{
	MacOS:   strPtr("linux"),
	Linux:   strPtr("linux"),
	Windows: strPtr("windows"),
}

// NewPlatformMap creates a new PlatformMap from a map of platform strings to values.
func NewPlatformMap[T any](values map[string]T) *PlatformMap[T] {
	p := &PlatformMap[T]{}
	for k, v := range values {
		val := v // capture value for pointer
		switch Platform(k) {
		case PlatformMacos:
			p.MacOS = &val
		case PlatformLinux:
			p.Linux = &val
		case PlatformWindows:
			p.Windows = &val
		default:
			panic(fmt.Sprintf("Unsupported platform key: %q", k))
		}
	}
	return p
}
