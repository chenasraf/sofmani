package appconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/eschao/config"
)

type AppConfig struct {
	Debug        bool               `json:"debug"          yaml:"debug"`
	CheckUpdates bool               `json:"check_updates"  yaml:"check_updates"`
	Install      []Installer        `json:"install"        yaml:"install"`
	Defaults     *AppConfigDefaults `json:"defaults"       yaml:"defaults"`
	Env          *map[string]string `json:"env"            yaml:"env"`
}

type AppCliConfig struct {
	ConfigFile   string
	Debug        *bool
	CheckUpdates *bool
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
	InstallerTypeGroup InstallerType = "group"
	InstallerTypeShell InstallerType = "shell"
	InstallerTypeBrew  InstallerType = "brew"
	InstallerTypeApt   InstallerType = "apt"
	InstallerTypeGit   InstallerType = "git"
	InstallerTypeRsync InstallerType = "rsync"
	InstallerTypeNpm   InstallerType = "npm"
	InstallerTypePnpm  InstallerType = "pnpm"
	InstallerTypeYarn  InstallerType = "yarn"
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

func ParseConfig() (*AppConfig, error) {
	overrides := ParseCliConfig()
	file := overrides.ConfigFile
	ext := filepath.Ext(file)
	switch ext {
	case ".json", ".yaml", ".yml":
		appConfig := AppConfig{}
		config.ParseConfigFile(&appConfig, file)
		if overrides.Debug != nil {
			appConfig.Debug = *overrides.Debug
		}
		if overrides.CheckUpdates != nil {
			appConfig.CheckUpdates = *overrides.CheckUpdates
		}
		return &appConfig, nil
	}
	return nil, fmt.Errorf("Unsupported config file extension %s", ext)
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

func ParseCliConfig() *AppCliConfig {
	args := os.Args[1:]
	config := &AppCliConfig{}
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
		case "-h", "--help":
			fmt.Println("Usage: sofmani [options] [config_file]")
			os.Exit(0)
		default:
			if file == "" {
				file = args[0]
			}
		}
		args = args[1:]
	}
	if file == "" {
		fmt.Println("No config file found")
		os.Exit(1)
	}
	config.ConfigFile = file
	return config
}
