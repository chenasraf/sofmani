package appconfig

import (
	"fmt"
	"path/filepath"

	"github.com/eschao/config"
)

type AppConfig struct {
	Debug        bool        `json:"debug"          yaml:"debug"`
	CheckUpdates bool        `json:"check_updates"  yaml:"check_updates"`
	Install      []Installer `json:"install"        yaml:"install"`
}

type Installer struct {
	Name      string          `json:"name"      yaml:"name"`
	Type      InstallerType   `json:"type"      yaml:"type"`
	Platforms *Platforms      `json:"platforms" yaml:"platforms"`
	Steps     *[]Installer    `json:"steps"     yaml:"steps"`
	Opts      *map[string]any `json:"opts" yaml:"opts"`
}

type InstallerType string

const (
	InstallerTypeCmd   InstallerType = "cmd"
	InstallerTypeBrew  InstallerType = "brew"
	InstallerTypeApt   InstallerType = "apt"
	InstallerTypeGit   InstallerType = "git"
	InstallerTypeGroup InstallerType = "group"
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
