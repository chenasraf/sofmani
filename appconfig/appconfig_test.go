package appconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chenasraf/sofmani/platform"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestPlatformMapResolve(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		expected *string
	}{
		{"MacOS", "darwin", lo.ToPtr("macos")},
		{"Linux", "linux", lo.ToPtr("linux")},
		{"Windows", "windows", lo.ToPtr("windows")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platform.SetOS(tt.platform)
			pm := platform.PlatformMap[string]{
				MacOS:   lo.ToPtr("macos"),
				Linux:   lo.ToPtr("linux"),
				Windows: lo.ToPtr("windows"),
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

func TestParseJsonConfig(t *testing.T) {
	// Create a temporary config file
	file, err := os.CreateTemp("", "config.*.json")
	assert.NoError(t, err)
	defer func() { assert.NoError(t, os.Remove(file.Name())) }()

	_, err = file.WriteString(`{"debug": true, "check_updates": false}`)
	assert.NoError(t, err)
	assert.NoError(t, file.Close())

	// Test parsing the config file
	overrides := AppCliConfig{ConfigFile: file.Name()}
	config, err := ParseConfig(&overrides)
	assert.NoError(t, err)
	assert.True(t, *config.Debug)
	assert.False(t, *config.CheckUpdates)
}

func TestParseYamlConfig(t *testing.T) {
	// Create a temporary config file
	file, err := os.CreateTemp("", "config.*.yaml")
	assert.NoError(t, err)
	defer func() { assert.NoError(t, os.Remove(file.Name())) }()

	_, err = file.WriteString(`
debug: true
check_updates: false
`)
	assert.NoError(t, err)
	assert.NoError(t, file.Close())

	// Test parsing the config file
	overrides := AppCliConfig{ConfigFile: file.Name()}
	config, err := ParseConfig(&overrides)
	assert.NoError(t, err)
	assert.True(t, *config.Debug)
	assert.False(t, *config.CheckUpdates)
}

func TestParseYamlConfigEnabled(t *testing.T) {
	// Create a temporary config file
	file, err := os.CreateTemp("", "config.*.yaml")
	assert.NoError(t, err)
	defer func() { assert.NoError(t, os.Remove(file.Name())) }()

	_, err = file.WriteString(`
debug: true
check_updates: false
install:
  - name: test
    type: shell
    enabled: true
`)
	assert.NoError(t, err)
	assert.NoError(t, file.Close())

	// Test parsing the config file
	overrides := AppCliConfig{ConfigFile: file.Name()}
	config, err := ParseConfig(&overrides)
	assert.NoError(t, err)
	assert.True(t, *config.Debug)
	assert.False(t, *config.CheckUpdates)
}

func TestGetRepoUpdateMode(t *testing.T) {
	t.Run("defaults to once when not configured", func(t *testing.T) {
		config := AppConfig{}
		assert.Equal(t, RepoUpdateOnce, config.GetRepoUpdateMode(InstallerTypeBrew))
		assert.Equal(t, RepoUpdateOnce, config.GetRepoUpdateMode(InstallerTypeApt))
	})

	t.Run("returns configured mode", func(t *testing.T) {
		repoUpdate := map[InstallerType]RepoUpdateMode{
			InstallerTypeBrew: RepoUpdateNever,
			InstallerTypeApt:  RepoUpdateAlways,
		}
		config := AppConfig{RepoUpdate: &repoUpdate}
		assert.Equal(t, RepoUpdateNever, config.GetRepoUpdateMode(InstallerTypeBrew))
		assert.Equal(t, RepoUpdateAlways, config.GetRepoUpdateMode(InstallerTypeApt))
	})

	t.Run("defaults to once for unconfigured type", func(t *testing.T) {
		repoUpdate := map[InstallerType]RepoUpdateMode{
			InstallerTypeBrew: RepoUpdateNever,
		}
		config := AppConfig{RepoUpdate: &repoUpdate}
		assert.Equal(t, RepoUpdateOnce, config.GetRepoUpdateMode(InstallerTypeApt))
	})

	t.Run("parses from yaml", func(t *testing.T) {
		file, err := os.CreateTemp("", "config.*.yaml")
		assert.NoError(t, err)
		defer func() { assert.NoError(t, os.Remove(file.Name())) }()

		_, err = file.WriteString(`
repo_update:
  brew: never
  apt: always
  apk: once
`)
		assert.NoError(t, err)
		assert.NoError(t, file.Close())

		config, err := ParseConfigFrom(file.Name())
		assert.NoError(t, err)
		assert.Equal(t, RepoUpdateNever, config.GetRepoUpdateMode(InstallerTypeBrew))
		assert.Equal(t, RepoUpdateAlways, config.GetRepoUpdateMode(InstallerTypeApt))
		assert.Equal(t, RepoUpdateOnce, config.GetRepoUpdateMode(InstallerTypeApk))
	})
}

func TestFindConfigFile(t *testing.T) {
	// Create a temporary config file
	dir := t.TempDir()
	file := filepath.Join(dir, "sofmani.json")
	err := os.WriteFile(file, []byte(`{"debug": true}`), 0644)
	assert.NoError(t, err)

	// Test finding the config file
	assert.NoError(t, os.Chdir(dir))
	assert.True(t, strings.HasSuffix(FindConfigFile(), file))
}
