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
	Name      string        `json:"name"      yaml:"name"`
	BinName   *string       `json:"bin_name"  yaml:"bin_name"`
	Type      InstallerType `json:"type"      yaml:"type"`
	Platforms *Platforms    `json:"platforms" yaml:"platforms"`
	Url       *string       `json:"url"       yaml:"url"`
	Command   *string       `json:"command"   yaml:"command"`
	Steps     *[]Installer  `json:"steps"     yaml:"steps"`
}

type InstallerType string

const (
	Brew  InstallerType = "brew"
	Apt   InstallerType = "apt"
	Git   InstallerType = "git"
	Cmd   InstallerType = "cmd"
	Group InstallerType = "group"
)

type Platforms struct {
	Only   *[]Platform `json:"only"   yaml:"only"`
	Except *[]Platform `json:"except" yaml:"except"`
}

type Platform string

const (
	Macos   Platform = "macos"
	Linux   Platform = "linux"
	Windows Platform = "windows"
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
