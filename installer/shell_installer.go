package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type ShellInstaller struct {
	InstallerBase
	Config *appconfig.AppConfig
	Info   *appconfig.InstallerData
}

type ShellOpts struct {
	Command       *string
	UpdateCommand *string
}

func (i *ShellInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	info := i.GetData()
	opts := i.GetOpts()
	if opts.Command == nil || len(*opts.Command) == 0 {
		errors = append(errors, ValidationError{FieldName: "command", Message: validationIsRequired(), InstallerName: *info.Name})
	}
	if opts.UpdateCommand != nil && len(*opts.UpdateCommand) == 0 {
		errors = append(errors, ValidationError{FieldName: "update_command", Message: validationIsRequired(), InstallerName: *info.Name})
	}
	return errors
}

// Install implements IInstaller.
func (i *ShellInstaller) Install() error {
	return i.RunCmdAsFile(*i.GetOpts().Command)
}

// Update implements IInstaller.
func (i *ShellInstaller) Update() error {
	if i.GetOpts().UpdateCommand != nil {
		return i.RunCmdAsFile(*i.GetOpts().UpdateCommand)
	}
	return i.Install()
}

// CheckNeedsUpdate implements IInstaller.
func (i *ShellInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	return false, nil
}

// CheckIsInstalled implements IInstaller.
func (i *ShellInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return i.RunCmdGetSuccess(utils.GetShellWhich(), i.GetBinName())
}

// GetData implements IInstaller.
func (i *ShellInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

func (i *ShellInstaller) GetOpts() *ShellOpts {
	opts := &ShellOpts{}
	info := i.Info
	if info.Opts != nil {
		if command, ok := (*info.Opts)["command"].(string); ok {
			opts.Command = &command
		}
		if updateCommand, ok := (*info.Opts)["update_command"].(string); ok {
			opts.UpdateCommand = &updateCommand
		}
	}
	return opts
}

func (i *ShellInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

func NewShellInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *ShellInstaller {
	return &ShellInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Info:          installer,
	}
}
