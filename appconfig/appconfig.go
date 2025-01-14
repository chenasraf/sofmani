package appconfig

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/chenasraf/sofmani/logger"
	"github.com/eschao/config"
)

type AppConfig struct {
	Debug        bool               `json:"debug"          yaml:"debug"`
	CheckUpdates bool               `json:"check_updates"  yaml:"check_updates"`
	Install      []Installer        `json:"install"        yaml:"install"`
	Defaults     *AppConfigDefaults `json:"defaults"       yaml:"defaults"`
	Env          *map[string]string `json:"env"            yaml:"env"`
	Filter       []string
}

type AppCliConfig struct {
	ConfigFile   string
	Debug        *bool
	CheckUpdates *bool
	Filter       []string
}

type AppConfigDefaults struct {
	Type *map[InstallerType]Installer `json:"type" yaml:"type"`
}

type Installer struct {
	Name           *string              `json:"name"              yaml:"name"`
	Type           InstallerType        `json:"type"              yaml:"type"`
	Env            *map[string]string   `json:"env"               yaml:"env"`
	Platforms      *Platforms           `json:"platforms"         yaml:"platforms"`
	Steps          *[]Installer         `json:"steps"             yaml:"steps"`
	Opts           *map[string]any      `json:"opts"              yaml:"opts"`
	BinName        *string              `json:"bin_name"          yaml:"bin_name"`
	CheckHasUpdate *string              `json:"check_has_update"  yaml:"check_has_update"`
	CheckInstalled *string              `json:"check_installed"   yaml:"check_installed"`
	PostInstall    *string              `json:"post_install"      yaml:"post_install"`
	PreInstall     *string              `json:"pre_install"       yaml:"pre_install"`
	PostUpdate     *string              `json:"post_update"       yaml:"post_update"`
	PreUpdate      *string              `json:"pre_update"        yaml:"pre_update"`
	EnvShell       *PlatformMap[string] `json:"env_shell"         yaml:"env_shell"`
}

type InstallerType string

const (
	InstallerTypeGroup    InstallerType = "group"
	InstallerTypeShell    InstallerType = "shell"
	InstallerTypeBrew     InstallerType = "brew"
	InstallerTypeApt      InstallerType = "apt"
	InstallerTypeGit      InstallerType = "git"
	InstallerTypeRsync    InstallerType = "rsync"
	InstallerTypeNpm      InstallerType = "npm"
	InstallerTypePnpm     InstallerType = "pnpm"
	InstallerTypeYarn     InstallerType = "yarn"
	InstallerTypeManifest InstallerType = "manifest"
)

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
	switch runtime.GOOS {
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

func (c *AppConfig) Environ() []string {
	if c.Env == nil {
		return []string{}
	}
	out := []string{}
	for k, v := range *c.Env {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}

func (i *Installer) Environ() []string {
	if i.Env == nil {
		return []string{}
	}
	out := []string{}
	for k, v := range *i.Env {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}

func ContainsPlatform(platforms *[]Platform, platform Platform) bool {
	for _, p := range *platforms {
		if p == platform {
			return true
		}
	}
	return false
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
			appConfig.Debug = *overrides.Debug
		}
		if overrides.CheckUpdates != nil {
			appConfig.CheckUpdates = *overrides.CheckUpdates
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
		Debug:        boolPtr(false),
		CheckUpdates: boolPtr(false),
		Filter:       []string{},
	}
	file := FindConfigFile()
	tVal := true
	fVal := false
	for len(args) > 0 {
		switch args[0] {
		case "-d", "--debug":
			config.Debug = &tVal
		case "-D", "--no-debug":
			config.Debug = &fVal
		case "-u", "--update":
			config.CheckUpdates = &tVal
		case "-U", "--no-update":
			config.CheckUpdates = &fVal
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
		Install: []Installer{},
	}
}
