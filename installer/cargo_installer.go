package installer

import (
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

// CargoInstaller is an installer for Rust cargo packages.
type CargoInstaller struct {
	InstallerBase
	// Config is the application configuration.
	Config *appconfig.AppConfig
	// Info is the installer data.
	Info *appconfig.InstallerData
}

// CargoOpts represents options for the CargoInstaller.
type CargoOpts struct {
	// Flags is a string of additional flags to pass to the cargo command.
	Flags *string
	// InstallFlags is a string of additional flags to pass only during install.
	InstallFlags *string
	// UpdateFlags is a string of additional flags to pass only during update.
	UpdateFlags *string
}

// Validate validates the installer configuration.
func (i *CargoInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	return errors
}

// Install implements IInstaller.
func (i *CargoInstaller) Install() error {
	name := *i.Info.Name
	opts := i.GetOpts()
	args := []string{"install"}
	if i.IsVerbose() {
		args = append(args, "--verbose")
	}
	if opts.InstallFlags != nil {
		args = append(args, strings.Fields(*opts.InstallFlags)...)
	} else if opts.Flags != nil {
		args = append(args, strings.Fields(*opts.Flags)...)
	}
	args = append(args, name)
	return i.RunCmdPassThrough("cargo", args...)
}

// Update implements IInstaller.
func (i *CargoInstaller) Update() error {
	name := *i.Info.Name
	opts := i.GetOpts()
	args := []string{"install"}
	if i.IsVerbose() {
		args = append(args, "--verbose")
	}
	if opts.UpdateFlags != nil {
		args = append(args, strings.Fields(*opts.UpdateFlags)...)
	} else if opts.Flags != nil {
		args = append(args, strings.Fields(*opts.Flags)...)
	}
	args = append(args, name)
	return i.RunCmdPassThrough("cargo", args...)
}

// CheckNeedsUpdate implements IInstaller.
func (i *CargoInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	// cargo install will skip if already up-to-date, so always attempt update
	return true, nil
}

// CheckIsInstalled implements IInstaller.
func (i *CargoInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return i.RunCmdGetSuccess(utils.GetShellWhich(), i.GetBinName())
}

// GetData implements IInstaller.
func (i *CargoInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

// GetOpts returns the parsed options for the CargoInstaller.
func (i *CargoInstaller) GetOpts() *CargoOpts {
	opts := &CargoOpts{}
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
func (i *CargoInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

// NewCargoInstaller creates a new CargoInstaller.
func NewCargoInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *CargoInstaller {
	i := &CargoInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Info:          installer,
	}

	return i
}
