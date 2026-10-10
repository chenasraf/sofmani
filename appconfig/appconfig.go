package appconfig

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/chenasraf/sofmani/platform"
	"github.com/chenasraf/sofmani/utils"
	"github.com/eschao/config"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

// CategoryDisplayMode controls how category headers are rendered.
type CategoryDisplayMode string

const (
	// CategoryDisplayBorder renders categories with a full border and spacing.
	CategoryDisplayBorder CategoryDisplayMode = "border"
	// CategoryDisplayBorderCompact renders categories with a border but no spacing.
	CategoryDisplayBorderCompact CategoryDisplayMode = "border-compact"
	// CategoryDisplayMinimal renders categories without a border or spacing.
	CategoryDisplayMinimal CategoryDisplayMode = "minimal"
)

// RepoUpdateMode controls how repository index updates are handled for a package manager.
type RepoUpdateMode string

const (
	// RepoUpdateOnce runs the repository update at most once per sofmani run (default).
	RepoUpdateOnce RepoUpdateMode = "once"
	// RepoUpdateAlways runs the repository update before every install/update check.
	RepoUpdateAlways RepoUpdateMode = "always"
	// RepoUpdateNever skips the repository update entirely.
	RepoUpdateNever RepoUpdateMode = "never"
)

// AppConfig represents the main application configuration.
type AppConfig struct {
	// Debug enables or disables debug mode.
	Debug *bool `json:"debug"          yaml:"debug"`
	// CheckUpdates enables or disables checking for updates.
	CheckUpdates *bool `json:"check_updates"  yaml:"check_updates"`
	// Summary enables or disables the installation summary at the end.
	Summary *bool `json:"summary"        yaml:"summary"`
	// CategoryDisplay controls how category headers are rendered.
	CategoryDisplay *CategoryDisplayMode `json:"category_display" yaml:"category_display"`
	// RepoUpdate controls repository index update behavior per installer type.
	// Supported types: brew, apt, apk. Values: "once" (default), "always", "never".
	RepoUpdate *map[InstallerType]RepoUpdateMode `json:"repo_update"    yaml:"repo_update"`
	// Install is a list of installers to run.
	Install []InstallerData `json:"install"        yaml:"install"`
	// Defaults provides default configurations for installer types.
	Defaults *AppConfigDefaults `json:"defaults"       yaml:"defaults"`
	// Env is a map of environment variables to set.
	Env *map[string]string `json:"env"            yaml:"env"`
	// PlatformEnv is a map of platform-specific environment variables to set.
	PlatformEnv *platform.PlatformMap[map[string]string] `json:"platform_env"   yaml:"platform_env"`
	// EnvCommand is a shell command, run once at startup, whose exported variables are added to
	// the environment of every installer. It loads secrets that are not exported in the shell
	// sofmani runs from.
	EnvCommand *EnvCommand `json:"env_command"    yaml:"env_command"`
	// LoadedEnv holds the variables EnvCommand exported. They are kept apart from Env because Env
	// values are resolved (~ and $VARS expanded) and these are secrets to pass on untouched.
	LoadedEnv map[string]string `json:"-"              yaml:"-"`
	// MachineAliases is a map of friendly names to machine IDs.
	MachineAliases *map[string]string `json:"machine_aliases" yaml:"machine_aliases"`
	// Filter is a list of installer names to filter by.
	Filter []string
	// IgnoreFrequency overrides frequency checks, running all installers regardless.
	IgnoreFrequency bool
	// StartFrom skips all installers before the one with the given name.
	StartFrom string
}

// GetRepoUpdateMode returns the repo update mode for the given installer type,
// defaulting to RepoUpdateOnce.
func (c *AppConfig) GetRepoUpdateMode(t InstallerType) RepoUpdateMode {
	if c.RepoUpdate != nil {
		if mode, ok := (*c.RepoUpdate)[t]; ok {
			return mode
		}
	}
	return RepoUpdateOnce
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
	// ShowVars indicates that the template variables and their current values should be shown.
	ShowVars bool
	// IgnoreFrequency overrides frequency checks, running all installers regardless.
	IgnoreFrequency bool
	// StartFrom skips all installers before the one with the given name.
	StartFrom string
}

// AppConfigDefaults provides default configurations for installer types.
type AppConfigDefaults struct {
	// Type is a map of installer types to their default configurations.
	Type *map[InstallerType]InstallerData `json:"type" yaml:"type"`
}

// GetCategoryDisplay returns the effective category display mode, defaulting to "border".
func (c *AppConfig) GetCategoryDisplay() CategoryDisplayMode {
	return lo.FromPtrOr(c.CategoryDisplay, CategoryDisplayBorder)
}

// Environ returns the combined environment variables as a slice of strings.
func (c *AppConfig) Environ() []string {
	return utils.EnvMapAsSlice(utils.CombineEnvMaps(c.Env, c.PlatformEnv.Resolve()))
}

// EnvCommand is a shell command that loads environment variables. In YAML it is either the
// command string or a map with `command`, `interactive` and `shell`.
type EnvCommand struct {
	// Command runs in the shell; the variables it exports are loaded.
	Command string `json:"command"     yaml:"command"`
	// Interactive starts the shell with -i, so its rc file (plugins, functions) loads first.
	Interactive bool `json:"interactive" yaml:"interactive"`
	// Shell overrides $SHELL.
	Shell string `json:"shell"       yaml:"shell"`
}

// envCommandFields has EnvCommand's fields without its unmarshalers, for decoding the map form.
type envCommandFields EnvCommand

// UnmarshalYAML accepts either the command string or the map form.
func (e *EnvCommand) UnmarshalYAML(unmarshal func(any) error) error {
	var command string
	if err := unmarshal(&command); err == nil {
		*e = EnvCommand{Command: command}
		return nil
	}
	var fields envCommandFields
	if err := unmarshal(&fields); err != nil {
		return err
	}
	*e = EnvCommand(fields)
	return nil
}

// UnmarshalJSON accepts either the command string or the map form.
func (e *EnvCommand) UnmarshalJSON(data []byte) error {
	var command string
	if err := json.Unmarshal(data, &command); err == nil {
		*e = EnvCommand{Command: command}
		return nil
	}
	var fields envCommandFields
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	*e = EnvCommand(fields)
	return nil
}

// LoadEnvCommand runs EnvCommand and stores the variables it exported in LoadedEnv, leaving out
// any that Env or PlatformEnv set, so configured values win. It returns the names of the
// variables loaded.
func (c *AppConfig) LoadEnvCommand() ([]string, error) {
	if c.EnvCommand == nil || c.EnvCommand.Command == "" {
		return nil, nil
	}
	shell := c.EnvCommand.Shell
	if shell == "" {
		shell = utils.GetOSShell(nil)
	}
	loaded, err := utils.RunEnvCommand(c.Environ(), shell, c.EnvCommand.Interactive, c.EnvCommand.Command)
	if err != nil {
		return nil, err
	}
	configured := utils.CombineEnvMaps(c.Env, c.PlatformEnv.Resolve())
	c.LoadedEnv = map[string]string{}
	for k, v := range loaded {
		if _, ok := configured[k]; !ok {
			c.LoadedEnv[k] = v
		}
	}
	names := slices.Sorted(maps.Keys(c.LoadedEnv))
	return names, nil
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
		appConfig.IgnoreFrequency = overrides.IgnoreFrequency
		appConfig.StartFrom = overrides.StartFrom
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
	desc = append(desc, fmt.Sprintf("Debug: %t", lo.FromPtrOr(c.Debug, false)))
	desc = append(desc, fmt.Sprintf("CheckUpdates: %t", lo.FromPtrOr(c.CheckUpdates, false)))
	desc = append(desc, fmt.Sprintf("Summary: %t", lo.FromPtrOr(c.Summary, true)))

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
			fmt.Fprintf(&filterBuilder, "\n  %s", f)
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
