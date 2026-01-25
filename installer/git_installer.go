package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

// GitInstaller is an installer for Git repositories.
type GitInstaller struct {
	InstallerBase
	// Config is the application configuration.
	Config *appconfig.AppConfig
	// Info is the installer data.
	Info *appconfig.InstallerData
}

// GitOpts represents options for the GitInstaller.
type GitOpts struct {
	// Destination is the directory where the repository will be cloned.
	Destination *string
	// Ref is the Git reference (branch, tag, or commit) to checkout.
	Ref *string
	// Flags is a string of additional flags to pass to git commands.
	Flags *string
	// InstallFlags is a string of additional flags to pass only to git clone.
	InstallFlags *string
	// UpdateFlags is a string of additional flags to pass only to git pull.
	UpdateFlags *string
}

// Validate validates the installer configuration.
func (i *GitInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	info := i.GetData()
	opts := i.GetOpts()
	if opts.Destination == nil || len(*opts.Destination) == 0 {
		errors = append(errors, ValidationError{FieldName: "destination", Message: validationIsRequired(), InstallerName: *info.Name})
	}
	if opts.Ref != nil && len(*opts.Ref) == 0 {
		errors = append(errors, ValidationError{FieldName: "ref", Message: validationIsNotEmpty(), InstallerName: *info.Name})
	}
	return errors
}

// Install implements IInstaller.
func (i *GitInstaller) Install() error {
	opts := i.GetOpts()
	args := []string{"clone"}
	if opts.InstallFlags != nil {
		args = append(args, strings.Fields(*opts.InstallFlags)...)
	} else if opts.Flags != nil {
		args = append(args, strings.Fields(*opts.Flags)...)
	}
	args = append(args, i.GetRepositoryUrl(), i.GetInstallDir())
	err := i.RunCmdPassThrough("git", args...)
	if err != nil {
		return err
	}
	if opts.Ref != nil {
		return i.RunCmdPassThrough("git", "-C", i.GetInstallDir(), "checkout", *opts.Ref)
	}
	return nil
}

// Update implements IInstaller.
func (i *GitInstaller) Update() error {
	opts := i.GetOpts()
	args := []string{"-C", i.GetInstallDir(), "pull"}
	if opts.UpdateFlags != nil {
		args = append(args, strings.Fields(*opts.UpdateFlags)...)
	} else if opts.Flags != nil {
		args = append(args, strings.Fields(*opts.Flags)...)
	}
	return i.RunCmdPassThrough("git", args...)
}

// CheckNeedsUpdate implements IInstaller.
func (i *GitInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	_, err := i.RunCmdGetSuccess("git", "-C", i.GetInstallDir(), "fetch")
	if err != nil {
		return false, err
	}
	output, err := i.RunCmdGetOutput("git", "-C", i.GetInstallDir(), "status", "-uno")
	if err != nil {
		return false, err
	}
	if strings.Contains(string(output), "Your branch is behind") {
		return true, nil
	}
	return false, nil
}

// CheckIsInstalled implements IInstaller.
func (i *GitInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return utils.PathExists(i.GetInstallDir())
}

// GetData implements IInstaller.
func (i *GitInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

// GetOpts returns the parsed options for the GitInstaller.
func (i *GitInstaller) GetOpts() *GitOpts {
	opts := &GitOpts{}
	info := i.Info
	if info.Opts != nil {
		if destination, ok := (*info.Opts)["destination"].(string); ok {
			destination = utils.GetRealPath(i.GetData().Environ(), destination)
			opts.Destination = &destination
		}
		if ref, ok := (*info.Opts)["ref"].(string); ok {
			opts.Ref = &ref
		}
		if flags, ok := (*info.Opts)["flags"].(string); ok {
			opts.Flags = &flags
		}
		if installFlags, ok := (*info.Opts)["install_flags"].(string); ok {
			opts.InstallFlags = &installFlags
		}
		if updateFlags, ok := (*info.Opts)["update_flags"].(string); ok {
			opts.UpdateFlags = &updateFlags
		}
	}
	return opts
}

// GetRepositoryUrl returns the URL of the Git repository.
// If the name in the installer data is a valid Git URL, it's returned directly.
// Otherwise, it's assumed to be a GitHub repository name (e.g., "owner/repo").
func (i *GitInstaller) GetRepositoryUrl() string {
	info := i.Info
	name := *info.Name
	if utils.IsGitURL(name) {
		return name
	}
	return fmt.Sprintf("https://github.com/%s", name)
}

// GetDestination returns the destination directory for the Git repository.
// It uses the Destination from the installer options if provided, otherwise it defaults to the current working directory.
func (i *GitInstaller) GetDestination() string {
	if i.GetOpts().Destination != nil {
		return *i.GetOpts().Destination
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}

// GetInstallDir returns the full path to the directory where the repository will be cloned.
// This is a combination of the destination directory and the base name of the repository.
func (i *GitInstaller) GetInstallDir() string {
	return filepath.Join(i.GetDestination(), filepath.Base(*i.Info.Name))
}

// NewGitInstaller creates a new GitInstaller.
func NewGitInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *GitInstaller {
	i := &GitInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Info:          installer,
	}

	return i
}
