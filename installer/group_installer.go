package installer

import (
	"fmt"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/summary"
	"github.com/chenasraf/sofmani/utils"
)

// GroupInstaller is an installer that groups other installers.
type GroupInstaller struct {
	InstallerBase
	// Config is the application configuration.
	Config *appconfig.AppConfig
	// Data is the installer data.
	Data *appconfig.InstallerData
	// childResults stores results from nested installers.
	childResults []summary.InstallResult
}

// GroupOpts represents options for the GroupInstaller.
type GroupOpts struct {
	//
}

// Validate validates the installer configuration.
func (i *GroupInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	info := i.GetData()
	if info.Steps == nil || len(*info.Steps) == 0 {
		errors = append(errors, ValidationError{FieldName: "steps", Message: "Must have at least one step", InstallerName: *info.Name})
	}
	return errors
}

// Install implements IInstaller.
func (i *GroupInstaller) Install() error {
	info := i.GetData()
	name := *info.Name
	logger.Debug("Installing group %s", name)
	i.childResults = []summary.InstallResult{}
	for _, step := range *i.Data.Steps {
		installer, err := GetInstaller(i.Config, &step)
		if err != nil {
			return err
		}
		if installer == nil {
			logger.Warn("Installer type %s is not supported, skipping", step.Type)
		} else {
			result, err := RunInstaller(i.Config, installer)
			if err != nil {
				logger.Error("Failed to run installer for step %s: %v", *step.Name, err)
				return fmt.Errorf("failed to run installer for step %s: %w", *step.Name, err)
			}
			if result != nil {
				i.childResults = append(i.childResults, *result)
			}
		}
	}
	return nil
}

// GetChildResults implements IChildResultsProvider.
func (i *GroupInstaller) GetChildResults() []summary.InstallResult {
	return i.childResults
}

// Update implements IInstaller.
func (i *GroupInstaller) Update() error {
	return i.Install()
}

// CheckNeedsUpdate implements IInstaller.
func (i *GroupInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	return true, nil
}

// CheckIsInstalled implements IInstaller.
func (i *GroupInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return i.RunCmdGetSuccess(utils.GetShellWhich(), i.GetBinName())
}

// GetData implements IInstaller.
func (i *GroupInstaller) GetData() *appconfig.InstallerData {
	return i.Data
}

// GetOpts returns the parsed options for the GroupInstaller.
func (i *GroupInstaller) GetOpts() *GroupOpts {
	return &GroupOpts{}
}

// GetBinName returns the binary name for the installer.
// It uses the BinName from the installer data if provided, otherwise it uses the installer name.
func (i *GroupInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

// NewGroupInstaller creates a new GroupInstaller.
func NewGroupInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *GroupInstaller {
	return &GroupInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Data:          installer,
	}
}
