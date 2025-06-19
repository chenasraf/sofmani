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
)

type AppConfig struct {
	Debug        *bool                                    `json:"debug"          yaml:"debug"`
	CheckUpdates *bool                                    `json:"check_updates"  yaml:"check_updates"`
	Install      []InstallerData                          `json:"install"        yaml:"install"`
	Defaults     *AppConfigDefaults                       `json:"defaults"       yaml:"defaults"`
	Env          *map[string]string                       `json:"env"            yaml:"env"`
	PlatformEnv  *platform.PlatformMap[map[string]string] `json:"platform_env"   yaml:"platform_env"`
	Filter       []string
}

type AppCliConfig struct {
	ConfigFile   string
	Debug        *bool
	CheckUpdates *bool
	Filter       []string
}

type AppConfigDefaults struct {
	Type *map[InstallerType]InstallerData `json:"type" yaml:"type"`
}

func (c *AppConfig) Environ() []string {
	return utils.EnvMapAsSlice(utils.CombineEnvMaps(c.Env, c.PlatformEnv.Resolve()))
}

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
	return nil, fmt.Errorf("Unsupported config file extension %s (filename: %s)", ext, file)
}

func ParseConfigFrom(file string) (*AppConfig, error) {
	appConfig := NewAppConfig()
	err := config.ParseConfigFile(&appConfig, file)
	if err != nil {
		return nil, err
	}
	return &appConfig, nil
}

func FindConfigFile() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	home, err := os.UserHomeDir()
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

func tryConfigDir(dir string) string {
	for _, ext := range []string{"json", "yaml", "yml"} {
		file := filepath.Join(dir, "sofmani."+ext)
		if _, err := os.Stat(file); err == nil {
			return file
		}
	}
	return ""
}

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

	filter := "Filter: "
	if len(c.Filter) > 0 {
		for _, f := range c.Filter {
			filter += fmt.Sprintf("\n  %s", f)
		}
	} else {
		filter += "None"
	}
	desc = append(desc, filter)

	return desc
}

var AppVersion string

func SetVersion(v string) {
	AppVersion = v
}

func boolPtr(b bool) *bool {
	return &b
}

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

func printVersion() {
	fmt.Println(AppVersion)
}

func NewAppConfig() AppConfig {
	return AppConfig{
		Install: []InstallerData{},
	}
}
