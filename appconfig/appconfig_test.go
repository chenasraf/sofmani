package appconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chenasraf/sofmani/platform"
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
			platform.SetOS(tt.platform)
			pm := platform.PlatformMap[string]{
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
	installer := InstallerData{Env: &env}
	expected := []string{"KEY1=value1", "KEY2=value2"}
	assert.ElementsMatch(t, expected, installer.Environ())
}

func TestInstallerPlatformEnviron(t *testing.T) {
	env := map[string]string{"KEY1": "value1", "KEY2": "value2"}
	platformEnv := map[string]string{"KEY2": "value2-override", "KEY3": "value3"}
	data := InstallerData{Env: &env, PlatformEnv: &platform.PlatformMap[map[string]string]{
		MacOS:   &platformEnv,
		Linux:   &platformEnv,
		Windows: &platformEnv,
	}}
	expected := []string{"KEY1=value1", "KEY2=value2-override", "KEY3=value3"}
	assert.ElementsMatch(t, expected, data.Environ())
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
	config, err := ParseConfig(&overrides)
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
