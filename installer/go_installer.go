package installer

import (
	"path"
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

// GoInstaller is an installer for Go packages installed via `go install`.
type GoInstaller struct {
	InstallerBase
	// Config is the application configuration.
	Config *appconfig.AppConfig
	// Info is the installer data.
	Info *appconfig.InstallerData
}

// GoOpts represents options for the GoInstaller.
type GoOpts struct {
	// Version is the module version to install (appended as `@version`).
	// Defaults to "latest" when neither this nor an inline `@version` on Name is set.
	Version *string
	// Flags is a string of additional flags to pass to the go install command.
	Flags *string
	// InstallFlags is a string of additional flags to pass only during install.
	InstallFlags *string
	// UpdateFlags is a string of additional flags to pass only during update.
	UpdateFlags *string
}

// Validate validates the installer configuration.
func (i *GoInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	return errors
}

// Install implements IInstaller.
func (i *GoInstaller) Install() error {
	opts := i.GetOpts()
	args := []string{"install"}
	if i.IsVerbose() {
		args = append(args, "-v")
	}
	if opts.InstallFlags != nil {
		args = append(args, strings.Fields(*opts.InstallFlags)...)
	} else if opts.Flags != nil {
		args = append(args, strings.Fields(*opts.Flags)...)
	}
	args = append(args, i.GetPackageRef())
	return i.RunCmdPassThrough("go", args...)
}

// Update implements IInstaller.
func (i *GoInstaller) Update() error {
	opts := i.GetOpts()
	args := []string{"install"}
	if i.IsVerbose() {
		args = append(args, "-v")
	}
	if opts.UpdateFlags != nil {
		args = append(args, strings.Fields(*opts.UpdateFlags)...)
	} else if opts.Flags != nil {
		args = append(args, strings.Fields(*opts.Flags)...)
	}
	args = append(args, i.GetPackageRef())
	return i.RunCmdPassThrough("go", args...)
}

// CheckNeedsUpdate implements IInstaller.
func (i *GoInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	// `go install pkg@latest` re-fetches and rebuilds only if newer; always attempt.
	return true, nil
}

// CheckIsInstalled implements IInstaller.
func (i *GoInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return i.RunCmdGetSuccess(utils.GetShellWhich(), i.GetBinName())
}

// GetData implements IInstaller.
func (i *GoInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

// GetOpts returns the parsed options for the GoInstaller.
func (i *GoInstaller) GetOpts() *GoOpts {
	opts := &GoOpts{}
	info := i.Info
	if info.Opts != nil {
		if version, ok := (*info.Opts)["version"].(string); ok {
			opts.Version = &version
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

// GetPackageRef returns the package reference to pass to `go install`,
// in the form `<pkg>@<version>`. If Name already contains an `@` the
// version is taken from there; otherwise opts.version (or "latest") is used.
func (i *GoInstaller) GetPackageRef() string {
	name := *i.Info.Name
	if strings.Contains(name, "@") {
		return name
	}
	version := "latest"
	if opts := i.GetOpts(); opts.Version != nil && *opts.Version != "" {
		version = *opts.Version
	}
	return name + "@" + version
}

// GetBinName returns the binary name for the installer.
// It uses the BinName from the installer data if provided, otherwise it
// derives one from the last path component of the package name.
func (i *GoInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	name := *info.Name
	if idx := strings.Index(name, "@"); idx >= 0 {
		name = name[:idx]
	}
	return path.Base(name)
}

// NewGoInstaller creates a new GoInstaller.
func NewGoInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *GoInstaller {
	i := &GoInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Info:          installer,
	}

	return i
}
