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
	// Version pins the package to an exact version, appended as `==version`.
	// Ignored when Name already carries a version specifier.
	Version *string
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
	args = append(args, i.GetPackageSpec())
	return i.RunCmdPassThrough("pipx", args...)
}

// Update implements IInstaller.
func (i *PipxInstaller) Update() error {
	opts := i.GetOpts()
	pinned := i.GetPinnedVersion()
	// `pipx upgrade` always moves to the newest release, so a pinned package is reinstalled
	// at its pinned version instead.
	args := []string{"upgrade"}
	if pinned != "" {
		args = []string{"install", "--force"}
	}
	if i.IsVerbose() {
		args = append(args, "--verbose")
	}
	if opts.UpdateFlags != nil {
		args = append(args, strings.Fields(*opts.UpdateFlags)...)
	} else if opts.Flags != nil {
		args = append(args, strings.Fields(*opts.Flags)...)
	}
	args = append(args, i.GetPackageSpec())
	return i.RunCmdPassThrough("pipx", args...)
}

// CheckNeedsUpdate implements IInstaller.
func (i *PipxInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	if pinned := i.GetPinnedVersion(); pinned != "" {
		return PinnedVersionNeedsUpdate(*i.Info.Name, pinned), nil
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

// GetPinnedVersion implements IVersionPinned.
func (i *PipxInstaller) GetPinnedVersion() string {
	if version := pipVersionSpec(*i.Info.Name); version != "" {
		return version
	}
	if version := i.GetOpts().Version; version != nil {
		return *version
	}
	return ""
}

// GetPackageSpec returns the package argument passed to pipx, as `<name>==<version>` when a
// version is pinned.
func (i *PipxInstaller) GetPackageSpec() string {
	name := *i.Info.Name
	if pipVersionSpec(name) != "" {
		return name
	}
	if version := i.GetPinnedVersion(); version != "" {
		return name + "==" + version
	}
	return name
}

// pipVersionSpec returns the version requirement of a pip package spec (`pkg==1.2.3`), or an
// empty string when the name carries none.
func pipVersionSpec(name string) string {
	if idx := strings.IndexAny(name, "=<>~!"); idx >= 0 {
		return strings.TrimLeft(name[idx:], "=<>~!")
	}
	return ""
}

// GetBinName returns the binary name for the installer.
// It uses the BinName from the installer data if provided, otherwise it uses the installer
// name with any version requirement stripped.
func (i *PipxInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	name := *info.Name
	if idx := strings.IndexAny(name, "=<>~!"); idx >= 0 {
		name = name[:idx]
	}
	return name
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
