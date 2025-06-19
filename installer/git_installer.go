package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type GitInstaller struct {
	InstallerBase
	Config *appconfig.AppConfig
	Info   *appconfig.InstallerData
}

type GitOpts struct {
	Destination *string
	Ref         *string
}

func (i *GitInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	info := i.GetData()
	opts := i.GetOpts()
	if opts.Destination == nil || len(*opts.Destination) == 0 {
		errors = append(errors, ValidationError{FieldName: "destination", Message: validationIsRequired(), InstallerName: *info.Name})
	}
	if opts.Ref == nil || len(*opts.Ref) == 0 {
		errors = append(errors, ValidationError{FieldName: "ref", Message: validationIsRequired(), InstallerName: *info.Name})
	}
	return errors
}

// Install implements IInstaller.
func (i *GitInstaller) Install() error {
	args := []string{"clone", i.GetRepositoryUrl(), i.GetInstallDir()}
	err := i.RunCmdPassThrough("git", args...)
	if err != nil {
		return err
	}
	if i.GetOpts().Ref != nil {
		return i.RunCmdPassThrough("git", "-C", i.GetInstallDir(), "checkout", *i.GetOpts().Ref)
	}
	return nil
}

// Update implements IInstaller.
func (i *GitInstaller) Update() error {
	return i.RunCmdPassThrough("git", "-C", i.GetInstallDir(), "pull")
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
	}
	return opts
}

func (i *GitInstaller) GetRepositoryUrl() string {
	info := i.Info
	name := *info.Name
	if utils.IsGitURL(name) {
		return name
	}
	return fmt.Sprintf("https://github.com/%s", name)
}

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

func (i *GitInstaller) GetInstallDir() string {
	return filepath.Join(i.GetDestination(), filepath.Base(*i.Info.Name))
}

func NewGitInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *GitInstaller {
	i := &GitInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Info:          installer,
	}

	return i
}
