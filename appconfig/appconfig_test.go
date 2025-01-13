package appconfig

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlatformMapResolve(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		expected *string
	}{
		{"MacOS", "darwin", strPtr("macos")},
		{"Linux", "linux", strPtr("linux")},
		{"Windows", "windows", strPtr("windows")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if runtime.GOOS != tt.platform {
				t.Skipf("Skipping test on %s", runtime.GOOS)
			}
			pm := PlatformMap[string]{
				MacOS:   strPtr("macos"),
				Linux:   strPtr("linux"),
				Windows: strPtr("windows"),
			}
			assert.Equal(t, tt.expected, pm.Resolve())
		})
	}
}

func TestAppConfigEnviron(t *testing.T) {
	env := map[string]string{"KEY1": "value1", "KEY2": "value2"}
	config := AppConfig{Env: &env}
	expected := []string{"KEY1=value1", "KEY2=value2"}
	assert.ElementsMatch(t, expected, config.Environ())
}

func TestInstallerEnviron(t *testing.T) {
	env := map[string]string{"KEY1": "value1", "KEY2": "value2"}
	installer := Installer{Env: &env}
	expected := []string{"KEY1=value1", "KEY2=value2"}
	assert.ElementsMatch(t, expected, installer.Environ())
}

func TestContainsPlatform(t *testing.T) {
	platforms := []Platform{PlatformMacos, PlatformLinux}
	assert.True(t, ContainsPlatform(&platforms, PlatformMacos))
	assert.False(t, ContainsPlatform(&platforms, PlatformWindows))
}

func TestParseConfig(t *testing.T) {
	// Create a temporary config file
	file, err := os.CreateTemp("", "config.*.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	_, err = file.WriteString(`{"debug": true, "check_updates": false}`)
	assert.NoError(t, err)
	file.Close()

	// Test parsing the config file
	overrides := AppCliConfig{ConfigFile: file.Name()}
	config, err := ParseConfig("1.0.0", &overrides)
	assert.NoError(t, err)
	assert.True(t, config.Debug)
	assert.False(t, config.CheckUpdates)
}

func TestFindConfigFile(t *testing.T) {
	// Create a temporary config file
	dir := t.TempDir()
	file := filepath.Join(dir, "sofmani.json")
	err := os.WriteFile(file, []byte(`{"debug": true}`), 0644)
	assert.NoError(t, err)

	// Test finding the config file
	os.Chdir(dir)
	assert.True(t, strings.HasSuffix(FindConfigFile(), file))
}

func strPtr(s string) *string {
	return &s
}
