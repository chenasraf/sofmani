package platform

import (
	"runtime"
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

func TestGetShouldRunOnOSEdgeCases(t *testing.T) {
	t.Run("nil Platforms returns true", func(t *testing.T) {
		var platforms *Platforms
		assert.True(t, platforms.GetShouldRunOnOS(PlatformMacos))
	})

	t.Run("empty Platforms returns true for any OS", func(t *testing.T) {
		platforms := Platforms{}
		assert.True(t, platforms.GetShouldRunOnOS(PlatformMacos))
		assert.True(t, platforms.GetShouldRunOnOS(PlatformLinux))
		assert.True(t, platforms.GetShouldRunOnOS(PlatformWindows))
	})

	t.Run("Only takes precedence over Except", func(t *testing.T) {
		// If both Only and Except are set, Only should take precedence
		platforms := Platforms{
			Only:   &[]Platform{PlatformMacos},
			Except: &[]Platform{PlatformMacos},
		}
		// Only should be checked first
		assert.True(t, platforms.GetShouldRunOnOS(PlatformMacos))
		assert.False(t, platforms.GetShouldRunOnOS(PlatformLinux))
	})

	t.Run("Multiple platforms in Only", func(t *testing.T) {
		platforms := Platforms{Only: &[]Platform{PlatformMacos, PlatformLinux}}
		assert.True(t, platforms.GetShouldRunOnOS(PlatformMacos))
		assert.True(t, platforms.GetShouldRunOnOS(PlatformLinux))
		assert.False(t, platforms.GetShouldRunOnOS(PlatformWindows))
	})

	t.Run("Multiple platforms in Except", func(t *testing.T) {
		platforms := Platforms{Except: &[]Platform{PlatformMacos, PlatformLinux}}
		assert.False(t, platforms.GetShouldRunOnOS(PlatformMacos))
		assert.False(t, platforms.GetShouldRunOnOS(PlatformLinux))
		assert.True(t, platforms.GetShouldRunOnOS(PlatformWindows))
	})
}

func TestGetArch(t *testing.T) {
	originalArch := archValue
	defer func() { SetArch(originalArch) }()

	t.Run("returns amd64 for amd64", func(t *testing.T) {
		SetArch("amd64")
		assert.Equal(t, ArchAmd64, GetArch())
	})

	t.Run("returns amd64 for x86_64", func(t *testing.T) {
		SetArch("x86_64")
		assert.Equal(t, ArchAmd64, GetArch())
	})

	t.Run("returns arm64 for arm64", func(t *testing.T) {
		SetArch("arm64")
		assert.Equal(t, ArchArm64, GetArch())
	})

	t.Run("returns arm64 for aarch64", func(t *testing.T) {
		SetArch("aarch64")
		assert.Equal(t, ArchArm64, GetArch())
	})

	t.Run("returns unknown arch as-is", func(t *testing.T) {
		SetArch("riscv64")
		assert.Equal(t, Architecture("riscv64"), GetArch())
	})
}

func TestGetArchAlias(t *testing.T) {
	originalArch := archValue
	defer func() { SetArch(originalArch) }()

	t.Run("returns x86_64 for amd64", func(t *testing.T) {
		SetArch("amd64")
		assert.Equal(t, "x86_64", GetArchAlias())
	})

	t.Run("returns arm64 for arm64", func(t *testing.T) {
		SetArch("arm64")
		assert.Equal(t, "arm64", GetArchAlias())
	})

	t.Run("returns unknown arch as-is", func(t *testing.T) {
		SetArch("riscv64")
		assert.Equal(t, "riscv64", GetArchAlias())
	})
}

func TestGetArchGnu(t *testing.T) {
	originalArch := archValue
	defer func() { SetArch(originalArch) }()

	t.Run("returns x86_64 for amd64", func(t *testing.T) {
		SetArch("amd64")
		assert.Equal(t, "x86_64", GetArchGnu())
	})

	t.Run("returns aarch64 for arm64", func(t *testing.T) {
		SetArch("arm64")
		assert.Equal(t, "aarch64", GetArchGnu())
	})

	t.Run("returns unknown arch as-is", func(t *testing.T) {
		SetArch("riscv64")
		assert.Equal(t, "riscv64", GetArchGnu())
	})
}

func TestPlatformMapResolve(t *testing.T) {
	originalOS := osValue
	defer func() { SetOS(originalOS) }()

	t.Run("returns nil for nil PlatformMap", func(t *testing.T) {
		var pm *PlatformMap[string]
		assert.Nil(t, pm.Resolve())
	})

	t.Run("returns MacOS value on darwin", func(t *testing.T) {
		SetOS("darwin")
		value := "mac-value"
		pm := &PlatformMap[string]{MacOS: &value}
		assert.Equal(t, &value, pm.Resolve())
	})

	t.Run("returns Linux value on linux", func(t *testing.T) {
		SetOS("linux")
		value := "linux-value"
		pm := &PlatformMap[string]{Linux: &value}
		assert.Equal(t, &value, pm.Resolve())
	})

	t.Run("returns Windows value on windows", func(t *testing.T) {
		SetOS("windows")
		value := "windows-value"
		pm := &PlatformMap[string]{Windows: &value}
		assert.Equal(t, &value, pm.Resolve())
	})

	t.Run("returns nil when platform value not set", func(t *testing.T) {
		SetOS("darwin")
		pm := &PlatformMap[string]{Linux: strPtr("linux-only")}
		assert.Nil(t, pm.Resolve())
	})

	t.Run("returns nil for unknown OS", func(t *testing.T) {
		SetOS("freebsd")
		pm := &PlatformMap[string]{
			MacOS:   strPtr("mac"),
			Linux:   strPtr("linux"),
			Windows: strPtr("windows"),
		}
		assert.Nil(t, pm.Resolve())
	})
}

func TestPlatformMapResolveWithFallback(t *testing.T) {
	originalOS := osValue
	defer func() { SetOS(originalOS) }()

	t.Run("returns primary value when set", func(t *testing.T) {
		SetOS("darwin")
		primary := &PlatformMap[string]{MacOS: strPtr("primary")}
		fallback := PlatformMap[string]{MacOS: strPtr("fallback")}
		assert.Equal(t, "primary", primary.ResolveWithFallback(fallback))
	})

	t.Run("returns fallback value when primary not set", func(t *testing.T) {
		SetOS("darwin")
		primary := &PlatformMap[string]{Linux: strPtr("linux-only")}
		fallback := PlatformMap[string]{MacOS: strPtr("fallback")}
		assert.Equal(t, "fallback", primary.ResolveWithFallback(fallback))
	})
}

func TestParsePlatformSingleValue(t *testing.T) {
	t.Run("sets value for all platforms", func(t *testing.T) {
		pm := ParsePlatformSingleValue("test-value")
		assert.NotNil(t, pm)
		assert.Equal(t, "test-value", *pm.MacOS)
		assert.Equal(t, "test-value", *pm.Linux)
		assert.Equal(t, "test-value", *pm.Windows)
	})
}

func TestParselatformMap(t *testing.T) {
	t.Run("parses macos value", func(t *testing.T) {
		values := map[string]string{"macos": "mac-value"}
		pm := ParselatformMap(values)
		assert.Equal(t, "mac-value", *pm.MacOS)
		assert.Nil(t, pm.Linux)
		assert.Nil(t, pm.Windows)
	})

	t.Run("parses linux value", func(t *testing.T) {
		values := map[string]string{"linux": "linux-value"}
		pm := ParselatformMap(values)
		assert.Nil(t, pm.MacOS)
		assert.Equal(t, "linux-value", *pm.Linux)
		assert.Nil(t, pm.Windows)
	})

	t.Run("parses windows value", func(t *testing.T) {
		values := map[string]string{"windows": "windows-value"}
		pm := ParselatformMap(values)
		assert.Nil(t, pm.MacOS)
		assert.Nil(t, pm.Linux)
		assert.Equal(t, "windows-value", *pm.Windows)
	})

	t.Run("parses all platforms", func(t *testing.T) {
		values := map[string]string{
			"macos":   "mac",
			"linux":   "lin",
			"windows": "win",
		}
		pm := ParselatformMap(values)
		assert.Equal(t, "mac", *pm.MacOS)
		assert.Equal(t, "lin", *pm.Linux)
		assert.Equal(t, "win", *pm.Windows)
	})
}

func TestNewPlatformMap(t *testing.T) {
	t.Run("returns nil for nil input", func(t *testing.T) {
		pm := NewPlatformMap[string](nil)
		assert.Nil(t, pm)
	})

	t.Run("handles single value", func(t *testing.T) {
		pm := NewPlatformMap[string]("single-value")
		assert.Equal(t, "single-value", *pm.MacOS)
		assert.Equal(t, "single-value", *pm.Linux)
		assert.Equal(t, "single-value", *pm.Windows)
	})

	t.Run("handles map of values", func(t *testing.T) {
		input := map[string]string{"macos": "mac", "linux": "lin"}
		pm := NewPlatformMap[string](input)
		assert.Equal(t, "mac", *pm.MacOS)
		assert.Equal(t, "lin", *pm.Linux)
		assert.Nil(t, pm.Windows)
	})

	t.Run("handles pointer to value", func(t *testing.T) {
		value := "ptr-value"
		pm := NewPlatformMap[string](&value)
		assert.Equal(t, "ptr-value", *pm.MacOS)
		assert.Equal(t, "ptr-value", *pm.Linux)
		assert.Equal(t, "ptr-value", *pm.Windows)
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		var ptr *string
		pm := NewPlatformMap[string](ptr)
		assert.Nil(t, pm)
	})

	t.Run("handles map of pointers", func(t *testing.T) {
		mac := "mac"
		input := map[string]*string{"macos": &mac, "linux": nil}
		pm := NewPlatformMap[string](input)
		assert.Equal(t, "mac", *pm.MacOS)
		assert.Nil(t, pm.Linux)
	})

	t.Run("handles map[any]any from YAML unmarshaling", func(t *testing.T) {
		// Simulate what YAML unmarshaling produces for nested maps
		input := map[any]any{
			"macos":   "mac-value",
			"linux":   "linux-value",
			"windows": "windows-value",
		}
		pm := NewPlatformMap[string](input)
		assert.Equal(t, "mac-value", *pm.MacOS)
		assert.Equal(t, "linux-value", *pm.Linux)
		assert.Equal(t, "windows-value", *pm.Windows)
	})

	t.Run("handles map[any]any with partial platforms", func(t *testing.T) {
		input := map[any]any{
			"macos": "mac-only",
		}
		pm := NewPlatformMap[string](input)
		assert.Equal(t, "mac-only", *pm.MacOS)
		assert.Nil(t, pm.Linux)
		assert.Nil(t, pm.Windows)
	})

	t.Run("handles map[string]any from JSON unmarshaling", func(t *testing.T) {
		// Simulate what JSON unmarshaling or mixed YAML produces
		input := map[string]any{
			"macos":   "mac-value",
			"linux":   "linux-value",
			"windows": "windows-value",
		}
		pm := NewPlatformMap[string](input)
		assert.Equal(t, "mac-value", *pm.MacOS)
		assert.Equal(t, "linux-value", *pm.Linux)
		assert.Equal(t, "windows-value", *pm.Windows)
	})

	t.Run("handles map[string]any with partial platforms", func(t *testing.T) {
		input := map[string]any{
			"linux": "linux-only",
		}
		pm := NewPlatformMap[string](input)
		assert.Nil(t, pm.MacOS)
		assert.Equal(t, "linux-only", *pm.Linux)
		assert.Nil(t, pm.Windows)
	})

	t.Run("handles map[any]any with non-string keys gracefully", func(t *testing.T) {
		// Non-string keys should be ignored
		input := map[any]any{
			"macos": "mac-value",
			123:     "ignored",
		}
		pm := NewPlatformMap[string](input)
		assert.Equal(t, "mac-value", *pm.MacOS)
		assert.Nil(t, pm.Linux)
		assert.Nil(t, pm.Windows)
	})

	t.Run("handles map[any]any with wrong value type gracefully", func(t *testing.T) {
		// Wrong value types should be ignored
		input := map[any]any{
			"macos": "mac-value",
			"linux": 123, // wrong type, should be ignored
		}
		pm := NewPlatformMap[string](input)
		assert.Equal(t, "mac-value", *pm.MacOS)
		assert.Nil(t, pm.Linux)
	})
}

func TestSetOSAndSetArch(t *testing.T) {
	t.Run("SetOS changes platform detection", func(t *testing.T) {
		originalOS := osValue
		defer func() { SetOS(originalOS) }()

		SetOS("darwin")
		assert.Equal(t, PlatformMacos, GetPlatform())

		SetOS("linux")
		assert.Equal(t, PlatformLinux, GetPlatform())

		SetOS("windows")
		assert.Equal(t, PlatformWindows, GetPlatform())
	})

	t.Run("SetArch changes architecture detection", func(t *testing.T) {
		originalArch := archValue
		defer func() { SetArch(originalArch) }()

		SetArch("amd64")
		assert.Equal(t, ArchAmd64, GetArch())

		SetArch("arm64")
		assert.Equal(t, ArchArm64, GetArch())
	})
}

func TestDockerOSMap(t *testing.T) {
	t.Run("has linux for MacOS", func(t *testing.T) {
		assert.Equal(t, "linux", *DockerOSMap.MacOS)
	})

	t.Run("has linux for Linux", func(t *testing.T) {
		assert.Equal(t, "linux", *DockerOSMap.Linux)
	})

	t.Run("has windows for Windows", func(t *testing.T) {
		assert.Equal(t, "windows", *DockerOSMap.Windows)
	})
}

func TestPlatformConstants(t *testing.T) {
	assert.Equal(t, Platform("macos"), PlatformMacos)
	assert.Equal(t, Platform("linux"), PlatformLinux)
	assert.Equal(t, Platform("windows"), PlatformWindows)
}

func TestArchConstants(t *testing.T) {
	assert.Equal(t, Architecture("amd64"), ArchAmd64)
	assert.Equal(t, Architecture("arm64"), ArchArm64)
}

func TestGetOSReturnsRuntimeValue(t *testing.T) {
	// Temporarily reset osValue to empty to test fallback
	originalOS := osValue
	defer func() { SetOS(originalOS) }()

	osValue = ""
	result := getOS()
	assert.Equal(t, runtime.GOOS, result)
}

func TestGetArchReturnsRuntimeValue(t *testing.T) {
	// Temporarily reset archValue to empty to test fallback
	originalArch := archValue
	defer func() { SetArch(originalArch) }()

	archValue = ""
	result := getArch()
	assert.Equal(t, runtime.GOARCH, result)
}
