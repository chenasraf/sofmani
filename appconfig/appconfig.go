package appconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eschao/config"
)

type AppConfig struct {
	Debug        bool               `json:"debug"          yaml:"debug"`
	CheckUpdates bool               `json:"check_updates"  yaml:"check_updates"`
	Install      []Installer        `json:"install"        yaml:"install"`
	Defaults     *AppConfigDefaults `json:"defaults"       yaml:"defaults"`
}

type AppConfigDefaults struct {
	Type *map[InstallerType]Installer `json:"type" yaml:"type"`
}

type Installer struct {
	Name           *string         `json:"name"              yaml:"name"`
	Type           InstallerType   `json:"type"              yaml:"type"`
	Platforms      *Platforms      `json:"platforms"         yaml:"platforms"`
	Steps          *[]Installer    `json:"steps"             yaml:"steps"`
	Opts           *map[string]any `json:"opts"              yaml:"opts"`
	BinName        *string         `json:"bin_name"          yaml:"bin_name"`
	CheckHasUpdate *string         `json:"check_has_update"  yaml:"check_has_update"`
	CheckInstalled *string         `json:"check_installed"   yaml:"check_installed"`
	PostInstall    *string         `json:"post_install"      yaml:"post_install"`
	PreInstall     *string         `json:"pre_install"       yaml:"pre_install"`
	PostUpdate     *string         `json:"post_update"       yaml:"post_update"`
	PreUpdate      *string         `json:"pre_update"        yaml:"pre_update"`
}

type InstallerType string

const (
	InstallerTypeGroup InstallerType = "group"
	InstallerTypeShell InstallerType = "shell"
	InstallerTypeBrew  InstallerType = "brew"
	InstallerTypeApt   InstallerType = "apt"
	InstallerTypeGit   InstallerType = "git"
	InstallerTypeRsync InstallerType = "rsync"
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

func ContainsPlatform(platforms *[]Platform, platform Platform) bool {
	for _, p := range *platforms {
		if p == platform {
			return true
		}
	}
	return false
}

func ParseConfigFile(file string) (*AppConfig, error) {
	ext := filepath.Ext(file)
	switch ext {
	case ".json", ".yaml", ".yml":
		appConfig := AppConfig{}
		config.ParseConfigFile(&appConfig, file)
		return &appConfig, nil
	}
	return nil, fmt.Errorf("Unsupported config file extension %s", ext)
}

func ApplyCliOverrides(config *AppConfig) *AppConfig {
	for len(os.Args) > 0 {
		switch os.Args[0] {
		case "-d", "--debug":
			config.Debug = true
		case "-D", "--no-debug":
			config.Debug = false
		case "-c", "--check-updates":
			config.CheckUpdates = true
		case "-C", "--no-check-updates":
			config.CheckUpdates = false
		}
		os.Args = os.Args[1:]
	}
	return config
}
