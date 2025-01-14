package appconfig

import (
	"fmt"

	"github.com/chenasraf/sofmani/platform"
)

type Installer struct {
	Name           *string                       `json:"name"              yaml:"name"`
	Type           InstallerType                 `json:"type"              yaml:"type"`
	Env            *map[string]string            `json:"env"               yaml:"env"`
	Platforms      *platform.Platforms           `json:"platforms"         yaml:"platforms"`
	Steps          *[]Installer                  `json:"steps"             yaml:"steps"`
	Opts           *map[string]any               `json:"opts"              yaml:"opts"`
	BinName        *string                       `json:"bin_name"          yaml:"bin_name"`
	CheckHasUpdate *string                       `json:"check_has_update"  yaml:"check_has_update"`
	CheckInstalled *string                       `json:"check_installed"   yaml:"check_installed"`
	PostInstall    *string                       `json:"post_install"      yaml:"post_install"`
	PreInstall     *string                       `json:"pre_install"       yaml:"pre_install"`
	PostUpdate     *string                       `json:"post_update"       yaml:"post_update"`
	PreUpdate      *string                       `json:"pre_update"        yaml:"pre_update"`
	EnvShell       *platform.PlatformMap[string] `json:"env_shell"         yaml:"env_shell"`
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
