package platform

import (
	"fmt"
	"runtime"
	"slices"
)

var osValue string = runtime.GOOS     // osValue stores the current operating system.
var archValue string = runtime.GOARCH // archValue stores the current architecture.

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

// getArch returns the current architecture. It caches the value after the first call.
func getArch() string {
	if archValue == "" {
		archValue = runtime.GOARCH
	}
	return archValue
}

// SetArch overrides the detected architecture. This is primarily used for testing.
func SetArch(v string) {
	archValue = v
}

// Architecture represents a CPU architecture.
type Architecture string

// Constants for supported architectures.
const (
	ArchAmd64 Architecture = "amd64" // ArchAmd64 represents x86_64 architecture.
	ArchArm64 Architecture = "arm64" // ArchArm64 represents ARM64 architecture.
)

// GetArch returns the current architecture (amd64 or arm64).
func GetArch() Architecture {
	switch getArch() {
	case "amd64", "x86_64":
		return ArchAmd64
	case "arm64", "aarch64":
		return ArchArm64
	default:
		return Architecture(getArch())
	}
}

// GetArchAlias returns the architecture in common alias format (x86_64 or arm64).
func GetArchAlias() string {
	switch GetArch() {
	case ArchAmd64:
		return "x86_64"
	case ArchArm64:
		return "arm64"
	default:
		return string(GetArch())
	}
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

// ParsePlatformSingleValue creates a new PlatformMap with the value for all platforms
func ParsePlatformSingleValue[T any](value T) *PlatformMap[T] {
	p := &PlatformMap[T]{}
	p.MacOS = &value
	p.Linux = &value
	p.Windows = &value
	return p
}

// ParselatformMap creates a PlatformMap from a map of platform strings to values.
func ParselatformMap[T any](values map[string]T) *PlatformMap[T] {
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

// NewPlatformMap creates a new PlatformMap from either a single value or a map.
func NewPlatformMap[T any](input any) *PlatformMap[T] {
	switch v := input.(type) {
	case nil:
		return nil
	case *T:
		if v != nil {
			return ParsePlatformSingleValue(*v)
		}
		return nil
	case map[string]T:
		return ParselatformMap(v)
	case map[string]*T:
		flat := make(map[string]T)
		for k, ptr := range v {
			if ptr != nil {
				flat[k] = *ptr
			}
		}
		return ParselatformMap(flat)
	case T:
		return ParsePlatformSingleValue(v)
	default:
		panic(fmt.Sprintf("NewPlatformMap: unsupported input type %T", input))
	}
}
