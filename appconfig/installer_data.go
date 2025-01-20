package appconfig

import (
	"strings"

	"github.com/chenasraf/sofmani/platform"
	"github.com/chenasraf/sofmani/utils"
	"github.com/samber/lo"
)

type InstallerData struct {
	Enabled        *string                                  `json:"enabled"           yaml:"enabled"`
	Name           *string                                  `json:"name"              yaml:"name"`
	Type           InstallerType                            `json:"type"              yaml:"type"`
	Tags           *string                                  `json:"tags"              yaml:"tags"`
	Env            *map[string]string                       `json:"env"               yaml:"env"`
	PlatformEnv    *platform.PlatformMap[map[string]string] `json:"platform_env"      yaml:"platform_env"`
	Platforms      *platform.Platforms                      `json:"platforms"         yaml:"platforms"`
	Steps          *[]InstallerData                         `json:"steps"             yaml:"steps"`
	Opts           *map[string]any                          `json:"opts"              yaml:"opts"`
	BinName        *string                                  `json:"bin_name"          yaml:"bin_name"`
	CheckHasUpdate *string                                  `json:"check_has_update"  yaml:"check_has_update"`
	CheckInstalled *string                                  `json:"check_installed"   yaml:"check_installed"`
	PostInstall    *string                                  `json:"post_install"      yaml:"post_install"`
	PreInstall     *string                                  `json:"pre_install"       yaml:"pre_install"`
	PostUpdate     *string                                  `json:"post_update"       yaml:"post_update"`
	PreUpdate      *string                                  `json:"pre_update"        yaml:"pre_update"`
	EnvShell       *platform.PlatformMap[string]            `json:"env_shell"         yaml:"env_shell"`
}

type InstallerType string

const (
	InstallerTypeGroup    InstallerType = "group"
	InstallerTypeShell    InstallerType = "shell"
	InstallerTypeBrew     InstallerType = "brew"
	InstallerTypeApt      InstallerType = "apt"
	InstallerTypeApk      InstallerType = "apk"
	InstallerTypeGit      InstallerType = "git"
	InstallerTypeRsync    InstallerType = "rsync"
	InstallerTypeNpm      InstallerType = "npm"
	InstallerTypePnpm     InstallerType = "pnpm"
	InstallerTypeYarn     InstallerType = "yarn"
	InstallerTypePipx     InstallerType = "pipx"
	InstallerTypeManifest InstallerType = "manifest"
)

func (i *InstallerData) Environ() []string {
	return utils.EnvMapAsSlice(utils.CombineEnvMaps(i.Env, i.PlatformEnv.Resolve()))
}

func (i *InstallerData) GetTagsList() []string {
	return lo.Map(strings.Split(*i.Tags, " "), func(tag string, i int) string {
		return strings.TrimSpace(tag)
	})
}
