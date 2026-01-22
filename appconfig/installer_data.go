package appconfig

import (
	"strings"

	"github.com/chenasraf/sofmani/machine"
	"github.com/chenasraf/sofmani/platform"
	"github.com/chenasraf/sofmani/utils"
	"github.com/samber/lo"
)

// SkipSummary controls whether an installer is excluded from the summary.
// It can be a boolean (applies to both install and update) or a map with
// "install" and "update" keys for granular control.
type SkipSummary struct {
	Install bool
	Update  bool
}

// UnmarshalYAML implements custom YAML unmarshaling for SkipSummary.
func (s *SkipSummary) UnmarshalYAML(unmarshal func(any) error) error {
	// Try boolean first
	var boolVal bool
	if err := unmarshal(&boolVal); err == nil {
		s.Install = boolVal
		s.Update = boolVal
		return nil
	}

	// Try map
	var mapVal map[string]bool
	if err := unmarshal(&mapVal); err == nil {
		if v, ok := mapVal["install"]; ok {
			s.Install = v
		}
		if v, ok := mapVal["update"]; ok {
			s.Update = v
		}
		return nil
	}

	return nil
}

// InstallerData represents the configuration for a single installer.
type InstallerData struct {
	// Enabled determines if the installer is enabled. Can be a boolean string ("true", "false") or a condition.
	Enabled *string `json:"enabled"           yaml:"enabled"`
	// Name is the name of the installer.
	Name *string `json:"name"              yaml:"name"`
	// Type is the type of the installer.
	Type InstallerType `json:"type"              yaml:"type"`
	// Tags is a space-separated list of tags for the installer.
	Tags *string `json:"tags"              yaml:"tags"`
	// Env is a map of environment variables to set for the installer.
	Env *map[string]string `json:"env"               yaml:"env"`
	// PlatformEnv is a map of platform-specific environment variables to set for the installer.
	PlatformEnv *platform.PlatformMap[map[string]string] `json:"platform_env"      yaml:"platform_env"`
	// Platforms is a list of platforms where this installer should run.
	Platforms *platform.Platforms `json:"platforms"         yaml:"platforms"`
	// Machines is a list of machine IDs where this installer should run.
	Machines *machine.Machines `json:"machines"          yaml:"machines"`
	// Steps is a list of sub-installers for group installers.
	Steps *[]InstallerData `json:"steps"             yaml:"steps"`
	// Opts is a map of options specific to the installer type.
	Opts *map[string]any `json:"opts"              yaml:"opts"`
	// BinName is the name of the binary to check for existence.
	BinName *string `json:"bin_name"          yaml:"bin_name"`
	// CheckHasUpdate is a command to check if an update is available.
	CheckHasUpdate *string `json:"check_has_update"  yaml:"check_has_update"`
	// CheckInstalled is a command to check if the software is installed.
	CheckInstalled *string `json:"check_installed"   yaml:"check_installed"`
	// PostInstall is a command to run after installation.
	PostInstall *string `json:"post_install"      yaml:"post_install"`
	// PreInstall is a command to run before installation.
	PreInstall *string `json:"pre_install"       yaml:"pre_install"`
	// PostUpdate is a command to run after updating.
	PostUpdate *string `json:"post_update"       yaml:"post_update"`
	// PreUpdate is a command to run before updating.
	PreUpdate *string `json:"pre_update"        yaml:"pre_update"`
	// EnvShell is a platform-specific shell to use for running commands.
	EnvShell *platform.PlatformMap[string] `json:"env_shell"         yaml:"env_shell"`
	// SkipSummary controls whether this installer is excluded from the summary.
	SkipSummary *SkipSummary `json:"skip_summary"      yaml:"skip_summary"`
}

// InstallerType represents the type of an installer.
type InstallerType string

// Constants for the different installer types.
const (
	InstallerTypeGroup         InstallerType = "group"          // InstallerTypeGroup represents a group of installers.
	InstallerTypeShell         InstallerType = "shell"          // InstallerTypeShell represents a shell command installer.
	InstallerTypeDocker        InstallerType = "docker"         // InstallerTypeDocker represents a Docker image installer.
	InstallerTypeBrew          InstallerType = "brew"           // InstallerTypeBrew represents a Homebrew package installer.
	InstallerTypeApt           InstallerType = "apt"            // InstallerTypeApt represents an APT package installer.
	InstallerTypeApk           InstallerType = "apk"            // InstallerTypeApk represents an APK package installer.
	InstallerTypeGit           InstallerType = "git"            // InstallerTypeGit represents a Git repository installer.
	InstallerTypeGitHubRelease InstallerType = "github-release" // InstallerTypeGitHubRelease represents a GitHub release installer.
	InstallerTypeRsync         InstallerType = "rsync"          // InstallerTypeRsync represents an rsync installer.
	InstallerTypeNpm           InstallerType = "npm"            // InstallerTypeNpm represents an npm package installer.
	InstallerTypePnpm          InstallerType = "pnpm"           // InstallerTypePnpm represents a pnpm package installer.
	InstallerTypeYarn          InstallerType = "yarn"           // InstallerTypeYarn represents a yarn package installer.
	InstallerTypePipx          InstallerType = "pipx"           // InstallerTypePipx represents a pipx package installer.
	InstallerTypeManifest      InstallerType = "manifest"       // InstallerTypeManifest represents a manifest file installer.
	InstallerTypePacman        InstallerType = "pacman"         // InstallerTypePacman represents a pacman package installer.
	InstallerTypeYay           InstallerType = "yay"            // InstallerTypeYay represents a yay (AUR helper) package installer.
)

// Environ returns the combined environment variables for the installer as a slice of strings.
func (i *InstallerData) Environ() []string {
	return utils.EnvMapAsSlice(utils.CombineEnvMaps(i.Env, i.PlatformEnv.Resolve()))
}

// GetTagsList returns the list of tags for the installer.
func (i *InstallerData) GetTagsList() []string {
	return lo.Map(strings.Split(*i.Tags, " "), func(tag string, i int) string {
		return strings.TrimSpace(tag)
	})
}
