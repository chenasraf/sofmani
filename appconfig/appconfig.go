package appconfig

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/chenasraf/sofmani/logger"
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
	// Install is a list of installers to run.
	Install []InstallerData `json:"install"        yaml:"install"`
	// Defaults provides default configurations for installer types.
	Defaults *AppConfigDefaults `json:"defaults"       yaml:"defaults"`
	// Env is a map of environment variables to set.
	Env *map[string]string `json:"env"            yaml:"env"`
	// PlatformEnv is a map of platform-specific environment variables to set.
	PlatformEnv *platform.PlatformMap[map[string]string] `json:"platform_env"   yaml:"platform_env"`
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
	// Filter is a list of installer names to filter by.
	Filter []string
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
		logger.Error("Failed to get user home directory: %v", err)
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
	desc = append(desc, fmt.Sprintf("Debug: %t", isDebug))
	desc = append(desc, fmt.Sprintf("CheckUpdates: %t", checkUpdates))

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

// boolPtr returns a pointer to a boolean value.
func boolPtr(b bool) *bool {
	return &b
}

// ParseCliConfig parses command-line arguments and returns an AppCliConfig.
func ParseCliConfig() *AppCliConfig {
	args := os.Args[1:]
	config := &AppCliConfig{
		ConfigFile:   "",
		Debug:        nil,
		CheckUpdates: nil,
		Filter:       []string{},
	}
	file := FindConfigFile()
	for len(args) > 0 {
		switch args[0] {
		case "-d", "--debug":
			config.Debug = boolPtr(true)
		case "-D", "--no-debug":
			config.Debug = boolPtr(false)
		case "-u", "--update":
			config.CheckUpdates = boolPtr(true)
		case "-U", "--no-update":
			config.CheckUpdates = boolPtr(false)
		case "-f", "--filter":
			if len(args) > 1 {
				config.Filter = append(config.Filter, args[1])
				args = args[1:]
			}
		case "-h", "--help":
			printHelp()
			os.Exit(0)
		case "-v", "--version":
			printVersion()
			os.Exit(0)
		default:
			if strings.HasPrefix(strings.TrimSpace(args[0]), "-test.") {
				break
			}
			_, err := os.Stat(file)
			exists := !errors.Is(err, fs.ErrNotExist)
			if exists {
				file = args[0]
			}
		}
		args = args[1:]
	}
	if file == "" {
		logger.Error("No config file found")
		os.Exit(1)
	}
	config.ConfigFile = file
	return config
}

// printHelp prints the command-line help message.
func printHelp() {
	fmt.Println("Usage: sofmani [options] [config_file]")
	fmt.Println("Options:")
	fmt.Println("  -d, --debug        Enable debug mode")
	fmt.Println("  -D, --no-debug     Disable debug mode")
	fmt.Println("  -u, --update       Enable update checks")
	fmt.Println("  -U, --no-update    Disable update checks")
	fmt.Println("  -h, --help         Show this help message")
	fmt.Println("  -f, --filter       Filter by installer name (can be used multiple times)")
	fmt.Println("  -v, --version      Show version")
	fmt.Println("")
	fmt.Println("For online documentation, see https://github.com/chenasraf/sofmani/tree/master/docs")
}

// printVersion prints the application version.
func printVersion() {
	fmt.Println(AppVersion)
}

// NewAppConfig creates a new AppConfig with default values.
func NewAppConfig() AppConfig {
	return AppConfig{
		Install: []InstallerData{},
	}
}
