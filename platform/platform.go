package platform

import (
	"fmt"
	"runtime"
	"slices"
)

var osValue string = runtime.GOOS

func getOS() string {
	if osValue == "" {
		osValue = runtime.GOOS
	}
	return osValue
}

func SetOS(v string) {
	osValue = v
}

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

type Platforms struct {
	Only   *[]Platform `json:"only"   yaml:"only"`
	Except *[]Platform `json:"except" yaml:"except"`
}

type Platform string

const (
	PlatformMacos   Platform = "macos"
	PlatformLinux   Platform = "linux"
	PlatformWindows Platform = "windows"
)

type PlatformMap[T any] struct {
	MacOS   *T `json:"macos"   yaml:"macos"`
	Linux   *T `json:"linux"   yaml:"linux"`
	Windows *T `json:"windows" yaml:"windows"`
}

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

func (o *PlatformMap[T]) ResolveWithFallback(fallback PlatformMap[T]) T {
	val := o.Resolve()
	if val == nil {
		return *fallback.Resolve()
	}
	return *val
}

func ContainsPlatform(platforms *[]Platform, platform Platform) bool {
	return slices.Contains(*platforms, platform)
}
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

func strPtr(s string) *string {
	return &s
}

var DockerArchMap = PlatformMap[string]{
	MacOS:   strPtr("amd64"),
	Linux:   strPtr("amd64"),
	Windows: strPtr("amd64"),
}

var DockerOSMap = PlatformMap[string]{
	MacOS:   strPtr("darwin"),
	Linux:   strPtr("linux"),
	Windows: strPtr("windows"),
}

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
