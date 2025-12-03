package installer

import (
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

// PipxInstaller is an installer for pipx packages.
type PipxInstaller struct {
	InstallerBase
	// Config is the application configuration.
	Config *appconfig.AppConfig
	// Info is the installer data.
	Info *appconfig.InstallerData
}

// PipxOpts represents options for the PipxInstaller.
type PipxOpts struct {
	// Flags is a string of additional flags to pass to the pipx command.
	Flags *string
	// InstallFlags is a string of additional flags to pass only during install.
	InstallFlags *string
	// UpdateFlags is a string of additional flags to pass only during update.
	UpdateFlags *string
}

// Validate validates the installer configuration.
func (i *PipxInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	return errors
}

// Install implements IInstaller.
func (i *PipxInstaller) Install() error {
	name := *i.Info.Name
	opts := i.GetOpts()
	args := []string{"install"}
	if opts.InstallFlags != nil {
		args = append(args, strings.Fields(*opts.InstallFlags)...)
	} else if opts.Flags != nil {
		args = append(args, strings.Fields(*opts.Flags)...)
	}
	args = append(args, name)
	return i.RunCmdPassThrough("pipx", args...)
}

// Update implements IInstaller.
func (i *PipxInstaller) Update() error {
	opts := i.GetOpts()
	args := []string{"upgrade"}
	if opts.UpdateFlags != nil {
		args = append(args, strings.Fields(*opts.UpdateFlags)...)
	} else if opts.Flags != nil {
		args = append(args, strings.Fields(*opts.Flags)...)
	}
	args = append(args, *i.Info.Name)
	return i.RunCmdPassThrough("pipx", args...)
}

// CheckNeedsUpdate implements IInstaller.
func (i *PipxInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	success, err := i.RunCmdGetSuccess("pipx", "upgrade", "--pip-args=--dry-run", *i.Info.Name)
	if err != nil {
		return false, err
	}
	return !success, nil
}

// CheckIsInstalled implements IInstaller.
func (i *PipxInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return i.RunCmdGetSuccess(utils.GetShellWhich(), i.GetBinName())
}

// GetData implements IInstaller.
func (i *PipxInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

// GetOpts returns the parsed options for the PipxInstaller.
func (i *PipxInstaller) GetOpts() *PipxOpts {
	opts := &PipxOpts{}
	info := i.Info
	if info.Opts != nil {
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

// GetBinName returns the binary name for the installer.
// It uses the BinName from the installer data if provided, otherwise it uses the installer name.
func (i *PipxInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

// NewPipxInstaller creates a new PipxInstaller.
func NewPipxInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *PipxInstaller {
	i := &PipxInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Info:          installer,
	}

	return i
}
