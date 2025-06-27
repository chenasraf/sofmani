package installer

import (
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/utils"
)

// RsyncInstaller is an installer that uses rsync to copy files.
type RsyncInstaller struct {
	InstallerBase
	// Config is the application configuration.
	Config *appconfig.AppConfig
	// Info is the installer data.
	Info *appconfig.InstallerData
}

// RsyncOpts represents options for the RsyncInstaller.
type RsyncOpts struct {
	// Source is the source directory or file.
	Source *string
	// Destination is the destination directory or file.
	Destination *string
	// Flags is a string of flags to pass to the rsync command.
	Flags *string
}

// Validate validates the installer configuration.
func (i *RsyncInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	info := i.GetData()
	opts := i.GetOpts()
	if opts.Source == nil || len(*opts.Source) == 0 {
		errors = append(errors, ValidationError{FieldName: "source", Message: validationIsRequired(), InstallerName: *info.Name})
	}
	if opts.Destination == nil || len(*opts.Destination) == 0 {
		errors = append(errors, ValidationError{FieldName: "destination", Message: validationIsRequired(), InstallerName: *info.Name})
	}
	if opts.Flags != nil && len(*opts.Flags) == 0 {
		errors = append(errors, ValidationError{FieldName: "flags", Message: validationIsNotEmpty(), InstallerName: *info.Name})
	}
	return errors
}

// Install implements IInstaller.
func (i *RsyncInstaller) Install() error {
	defaultFlags := "-tr"
	if i.Config.Debug != nil && *i.Config.Debug {
		defaultFlags += "v"
	}
	flags := []string{defaultFlags}
	if i.GetOpts().Flags != nil {
		for _, flag := range strings.Split(*i.GetOpts().Flags, " ") {
			flags = append(flags, flag)
		}
	}
	data := i.GetData()
	env := data.Environ()
	src := utils.GetRealPath(env, *i.GetOpts().Source)
	dest := utils.GetRealPath(env, *i.GetOpts().Destination)

	flags = append(flags, src)
	flags = append(flags, dest)

	logger.Debug("rsync %s to %s", src, dest)
	return i.RunCmdPassThrough("rsync", flags...)
}

// Update implements IInstaller.
func (i *RsyncInstaller) Update() error {
	return i.Install()
}

// CheckNeedsUpdate implements IInstaller.
func (i *RsyncInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	return true, nil
}

// CheckIsInstalled implements IInstaller.
func (i *RsyncInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return false, nil
}

// GetData implements IInstaller.
func (i *RsyncInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

// GetOpts returns the parsed options for the RsyncInstaller.
func (i *RsyncInstaller) GetOpts() *RsyncOpts {
	opts := &RsyncOpts{}
	info := i.Info
	if info.Opts != nil {
		if src, ok := (*info.Opts)["source"].(string); ok {
			opts.Source = &src
		}
		if dest, ok := (*info.Opts)["destination"].(string); ok {
			opts.Destination = &dest
		}
		if flags, ok := (*info.Opts)["flags"].(string); ok {
			opts.Flags = &flags
		}
	}
	return opts
}

// GetBinName returns the binary name for the installer.
// For rsync, this is typically not applicable as it's a file transfer, not a binary installation.
// It defaults to the installer name.
func (i *RsyncInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

// NewRsyncInstaller creates a new RsyncInstaller.
func NewRsyncInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *RsyncInstaller {
	i := &RsyncInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Info:          installer,
	}

	return i
}
