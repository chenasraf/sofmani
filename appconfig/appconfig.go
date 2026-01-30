package appconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chenasraf/sofmani/platform"
	"github.com/chenasraf/sofmani/utils"
	"github.com/eschao/config"
	"gopkg.in/yaml.v3"
)

// AppConfig represents the main application configuration.
type AppConfig struct {
	// Debug enables or disables debug mode.
	Debug *bool `json:"debug"          yaml:"debug"`
	// CheckUpdates enables or disables checking for updates.
	CheckUpdates *bool `json:"check_updates"  yaml:"check_updates"`
	// Summary enables or disables the installation summary at the end.
	Summary *bool `json:"summary"        yaml:"summary"`
	// Install is a list of installers to run.
	Install []InstallerData `json:"install"        yaml:"install"`
	// Defaults provides default configurations for installer types.
	Defaults *AppConfigDefaults `json:"defaults"       yaml:"defaults"`
	// Env is a map of environment variables to set.
	Env *map[string]string `json:"env"            yaml:"env"`
	// PlatformEnv is a map of platform-specific environment variables to set.
	PlatformEnv *platform.PlatformMap[map[string]string] `json:"platform_env"   yaml:"platform_env"`
	// MachineAliases is a map of friendly names to machine IDs.
	MachineAliases *map[string]string `json:"machine_aliases" yaml:"machine_aliases"`
	// Filter is a list of installer names to filter by.
	Filter []string
}

// AppCliConfig represents the command-line interface configuration.
type AppCliConfig struct {
	// ConfigFile is the path to the configuration file.
	ConfigFile string
	// Debug enables or disables debug mode.
	Debug *bool
	// CheckUpdates enables or disables checking for updates.
	CheckUpdates *bool
	// Summary enables or disables the installation summary at the end.
	Summary *bool
	// Filter is a list of installer names to filter by.
	Filter []string
	// LogFile is the path to the log file.
	LogFile *string
	// ShowLogFile indicates that only the log file path should be shown.
	ShowLogFile bool
	// ShowMachineID indicates that only the machine ID should be shown.
	ShowMachineID bool
}

// AppConfigDefaults provides default configurations for installer types.
type AppConfigDefaults struct {
	// Type is a map of installer types to their default configurations.
	Type *map[InstallerType]InstallerData `json:"type" yaml:"type"`
}

// Environ returns the combined environment variables as a slice of strings.
func (c *AppConfig) Environ() []string {
	return utils.EnvMapAsSlice(utils.CombineEnvMaps(c.Env, c.PlatformEnv.Resolve()))
}

// ParseConfig parses the configuration file and applies overrides.
func ParseConfig(overrides *AppCliConfig) (*AppConfig, error) {
	file := overrides.ConfigFile
	ext := filepath.Ext(file)
	switch ext {
	case ".json", ".yaml", ".yml":
		appConfig, err := ParseConfigFrom(file)
		if err != nil {
			return nil, err
		}
		if overrides.Debug != nil {
			appConfig.Debug = overrides.Debug
		}
		if overrides.CheckUpdates != nil {
			appConfig.CheckUpdates = overrides.CheckUpdates
		}
		if overrides.Summary != nil {
			appConfig.Summary = overrides.Summary
		}
		appConfig.Filter = overrides.Filter
		return appConfig, nil
	}
	return nil, fmt.Errorf("unsupported config file extension %s (filename: %s)", ext, file)
}

// ParseConfigFrom parses the configuration from the given file.
func ParseConfigFrom(file string) (*AppConfig, error) {
	appConfig := NewAppConfig()
	err := config.ParseConfigFile(&appConfig, file)
	if err != nil {
		return nil, err
	}
	return &appConfig, nil
}

// ParseConfigFromContent parses the configuration from YAML content.
func ParseConfigFromContent(content []byte) (*AppConfig, error) {
	appConfig := NewAppConfig()
	err := yaml.Unmarshal(content, &appConfig)
	if err != nil {
		return nil, err
	}
	return &appConfig, nil
}

// FindConfigFile searches for the configuration file in standard locations.
// It searches in the current working directory, then in ~/.config, and finally in the home directory.
// It returns the path to the first file found, or an empty string if no file is found.
func FindConfigFile() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get user home directory: %v\n", err)
		return ""
	}
	file := ""
	dirs := []string{wd, filepath.Join(home, ".config"), home}
	for _, dir := range dirs {
		file = tryConfigDir(dir)
		if file != "" {
			return file
		}
	}
	return ""
}

// tryConfigDir attempts to find a configuration file with a valid extension in the given directory.
// It checks for "sofmani.json", "sofmani.yaml", and "sofmani.yml".
// It returns the path to the first file found, or an empty string if no file is found.
func tryConfigDir(dir string) string {
	for _, ext := range []string{"json", "yaml", "yml"} {
		file := filepath.Join(dir, "sofmani."+ext)
		if _, err := os.Stat(file); err == nil {
			return file
		}
	}
	return ""
}

// GetConfigDesc returns a string slice describing the current configuration.
func (c *AppConfig) GetConfigDesc() []string {
	desc := []string{}
	isDebug := false
	if c.Debug != nil {
		isDebug = *c.Debug
	}
	checkUpdates := false
	if c.CheckUpdates != nil {
		checkUpdates = *c.CheckUpdates
	}
	showSummary := true // default is enabled
	if c.Summary != nil {
		showSummary = *c.Summary
	}
	desc = append(desc, fmt.Sprintf("Debug: %t", isDebug))
	desc = append(desc, fmt.Sprintf("CheckUpdates: %t", checkUpdates))
	desc = append(desc, fmt.Sprintf("Summary: %t", showSummary))

	if c.Env != nil {
		desc = append(desc, "Environment Variables:")
		for k, v := range *c.Env {
			desc = append(desc, fmt.Sprintf("  %s=%s", k, v))
		}
	}

	if c.PlatformEnv != nil {
		desc = append(desc, "Platform Environment Variables:\n")
		desc = append(desc, fmt.Sprintf("  %s", platform.GetPlatform()))
		for k, v := range *c.PlatformEnv.Resolve() {
			desc = append(desc, fmt.Sprintf("  %s=%s", k, v))
		}
	}

	var filterBuilder strings.Builder
	filterBuilder.WriteString("Filter: ")
	if len(c.Filter) > 0 {
		for _, f := range c.Filter {
			filterBuilder.WriteString(fmt.Sprintf("\n  %s", f))
		}
	} else {
		filterBuilder.WriteString("None")
	}
	desc = append(desc, filterBuilder.String())

	return desc
}

// AppVersion is the current version of the application.
var AppVersion string

// SetVersion sets the application version.
func SetVersion(v string) {
	AppVersion = v
}

// NewAppConfig creates a new AppConfig with default values.
func NewAppConfig() AppConfig {
	return AppConfig{
		Install: []InstallerData{},
	}
}
