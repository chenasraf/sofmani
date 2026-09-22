package installer

import (
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

// NpmInstaller is an installer for npm, pnpm, and yarn packages.
type NpmInstaller struct {
	InstallerBase
	// Config is the application configuration.
	Config *appconfig.AppConfig
	// PackageManager is the package manager to use (npm, pnpm, or yarn).
	PackageManager NpmPackageManager
	// Info is the installer data.
	Info *appconfig.InstallerData
}

// NpmOpts represents options for the NpmInstaller.
type NpmOpts struct {
	// Version pins the package to an exact version, appended as `@version`.
	// Ignored when Name already carries a `@version` suffix.
	Version *string
	// Flags is a string of additional flags to pass to the npm/pnpm/yarn command.
	Flags *string
	// InstallFlags is a string of additional flags to pass only during install.
	InstallFlags *string
	// UpdateFlags is a string of additional flags to pass only during update.
	UpdateFlags *string
}

// NpmPackageManager represents a Node.js package manager type.
// This type is also defined in apt_installer.go. Consider refactoring to a common location if appropriate.
type NpmPackageManager string

// Constants for supported Node.js package managers.
const (
	PackageManagerNpm  NpmPackageManager = "npm"  // PackageManagerNpm represents the npm package manager.
	PackageManagerYarn NpmPackageManager = "yarn" // PackageManagerYarn represents the yarn package manager.
	PackageManagerPnpm NpmPackageManager = "pnpm" // PackageManagerPnpm represents the pnpm package manager.
)

// Validate validates the installer configuration.
func (i *NpmInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	return errors
}

// Install implements IInstaller.
func (i *NpmInstaller) Install() error {
	opts := i.GetOpts()
	args := []string{"install", "--global"}
	if i.IsVerbose() {
		args = append(args, "--verbose")
	}
	if opts.InstallFlags != nil {
		args = append(args, strings.Fields(*opts.InstallFlags)...)
	} else if opts.Flags != nil {
		args = append(args, strings.Fields(*opts.Flags)...)
	}
	args = append(args, i.GetPackageSpec())
	return i.RunCmdPassThrough(string(i.PackageManager), args...)
}

// Update implements IInstaller.
func (i *NpmInstaller) Update() error {
	opts := i.GetOpts()
	args := []string{"install", "--global"}
	if i.IsVerbose() {
		args = append(args, "--verbose")
	}
	if opts.UpdateFlags != nil {
		args = append(args, strings.Fields(*opts.UpdateFlags)...)
	} else if opts.Flags != nil {
		args = append(args, strings.Fields(*opts.Flags)...)
	}
	target := *i.Info.Name + "@latest"
	if i.GetPinnedVersion() != "" {
		target = i.GetPackageSpec()
	}
	args = append(args, target)
	return i.RunCmdPassThrough(string(i.PackageManager), args...)
}

// CheckNeedsUpdate implements IInstaller.
func (i *NpmInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	// A pinned package is always "outdated" as far as the registry is concerned, so the pin
	// itself decides instead.
	if pinned := i.GetPinnedVersion(); pinned != "" {
		return PinnedVersionNeedsUpdate(*i.Info.Name, pinned), nil
	}
	success, err := i.RunCmdGetSuccess(string(i.PackageManager), "outdated", "--global", "--json", *i.Info.Name)
	if err != nil {
		return false, err
	}
	return !success, nil
}

// CheckIsInstalled implements IInstaller.
func (i *NpmInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return i.RunCmdGetSuccess(utils.GetShellWhich(), i.GetBinName())
}

// GetData implements IInstaller.
func (i *NpmInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

// GetOpts returns the parsed options for the NpmInstaller.
func (i *NpmInstaller) GetOpts() *NpmOpts {
	opts := &NpmOpts{}
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

// GetPinnedVersion implements IVersionPinned. The version may come from opts.version or
// from a `@version` suffix on the package name.
func (i *NpmInstaller) GetPinnedVersion() string {
	if version := npmVersionSpec(*i.Info.Name); version != "" {
		return version
	}
	if version := i.GetOpts().Version; version != nil {
		return *version
	}
	return ""
}

// GetPackageSpec returns the package argument passed to the package manager, as
// `<name>@<version>` when a version is pinned.
func (i *NpmInstaller) GetPackageSpec() string {
	name := *i.Info.Name
	if npmVersionSpec(name) != "" {
		return name
	}
	if version := i.GetPinnedVersion(); version != "" {
		return name + "@" + version
	}
	return name
}

// npmVersionSpec returns the version part of an npm package spec (`pkg@1.2.3`), or an empty
// string when the name carries no version. A leading `@` belongs to the package scope.
func npmVersionSpec(name string) string {
	if idx := strings.LastIndex(name, "@"); idx > 0 {
		return name[idx+1:]
	}
	return ""
}

// GetBinName returns the binary name for the installer.
// It uses the BinName from the installer data if provided, otherwise it uses the installer
// name with any `@version` suffix stripped.
func (i *NpmInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	name := *info.Name
	if idx := strings.LastIndex(name, "@"); idx > 0 {
		name = name[:idx]
	}
	return name
}

// NewNpmInstaller creates a new NpmInstaller.
func NewNpmInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *NpmInstaller {
	var packageManager NpmPackageManager
	switch installer.Type {
	case appconfig.InstallerTypeNpm:
		packageManager = PackageManagerNpm
	case appconfig.InstallerTypePnpm:
		packageManager = PackageManagerPnpm
	case appconfig.InstallerTypeYarn:
		packageManager = PackageManagerYarn
	}
	i := &NpmInstaller{
		InstallerBase:  InstallerBase{Data: installer},
		Config:         cfg,
		PackageManager: packageManager,
		Info:           installer,
	}

	return i
}
