package appconfig

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chenasraf/sofmani/logger"
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

func TestLoadEnvCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX shell command")
	}
	platform.SetOS(runtime.GOOS)
	logger.InitLogger(false)

	t.Run("loads exported variables, keeping configured ones out", func(t *testing.T) {
		cfg := &AppConfig{
			Env:        &map[string]string{"KEPT": "config", "VAULT": "work"},
			EnvCommand: &EnvCommand{Command: `export KEPT=loaded; export TOKEN="secret-$VAULT"`, Shell: "sh"},
		}
		names, err := cfg.LoadEnvCommand()
		assert.NoError(t, err)
		assert.Equal(t, []string{"TOKEN"}, names)
		assert.Equal(t, map[string]string{"TOKEN": "secret-work"}, cfg.LoadedEnv)
		assert.Equal(t, map[string]string{"KEPT": "config", "VAULT": "work"}, *cfg.Env)
	})

	t.Run("does nothing without a command", func(t *testing.T) {
		cfg := &AppConfig{}
		names, err := cfg.LoadEnvCommand()
		assert.NoError(t, err)
		assert.Empty(t, names)
		assert.Nil(t, cfg.LoadedEnv)
	})

	t.Run("reports a failing command", func(t *testing.T) {
		cfg := &AppConfig{EnvCommand: &EnvCommand{Command: "exit 1", Shell: "sh"}}
		_, err := cfg.LoadEnvCommand()
		assert.Error(t, err)
	})
}

func TestEnvCommandParsing(t *testing.T) {
	t.Run("string form in yaml", func(t *testing.T) {
		cfg, err := ParseConfigFromContent([]byte("env_command: load-secrets\ninstall: []\n"))
		assert.NoError(t, err)
		assert.Equal(t, &EnvCommand{Command: "load-secrets"}, cfg.EnvCommand)
	})

	t.Run("map form in yaml", func(t *testing.T) {
		cfg, err := ParseConfigFromContent([]byte("env_command:\n  command: load_secrets\n  interactive: true\n  shell: zsh\ninstall: []\n"))
		assert.NoError(t, err)
		assert.Equal(t, &EnvCommand{Command: "load_secrets", Interactive: true, Shell: "zsh"}, cfg.EnvCommand)
	})

	t.Run("both forms in json", func(t *testing.T) {
		dir := t.TempDir()
		for content, want := range map[string]EnvCommand{
			`{"env_command": "load-secrets", "install": []}`:                                   {Command: "load-secrets"},
			`{"env_command": {"command": "load_secrets", "interactive": true}, "install": []}`: {Command: "load_secrets", Interactive: true},
		} {
			file := filepath.Join(dir, "sofmani.json")
			assert.NoError(t, os.WriteFile(file, []byte(content), 0644))
			cfg, err := ParseConfigFrom(file)
			assert.NoError(t, err)
			assert.Equal(t, &want, cfg.EnvCommand)
		}
	})

	t.Run("map form from a yaml file", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "sofmani.yml")
		assert.NoError(t, os.WriteFile(file, []byte("env_command:\n  command: load_secrets\n  interactive: true\ninstall: []\n"), 0644))
		cfg, err := ParseConfigFrom(file)
		assert.NoError(t, err)
		assert.Equal(t, &EnvCommand{Command: "load_secrets", Interactive: true}, cfg.EnvCommand)
	})
}
