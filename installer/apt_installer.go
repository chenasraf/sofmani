package installer

import (
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

// AptInstaller is an installer for apt and apk packages.
type AptInstaller struct {
	InstallerBase
	// Config is the application configuration.
	Config *appconfig.AppConfig
	// Info is the installer data.
	Info *appconfig.InstallerData
	// PackageManager is the package manager to use (apt or apk).
	PackageManager AptPackageManager
}

// AptOpts represents options for the AptInstaller.
type AptOpts struct {
	// Version pins the package to an exact version, appended as `=version`.
	// Ignored when Name already carries a `=version` suffix.
	Version *string
	// Flags is a string of additional flags to pass to the apt/apk command.
	Flags *string
	// InstallFlags is a string of additional flags to pass only during install.
	InstallFlags *string
	// UpdateFlags is a string of additional flags to pass only during update.
	UpdateFlags *string
}

// AptPackageManager represents a package manager type.
type AptPackageManager string

// Constants for supported package managers.
const (
	PackageManagerApk AptPackageManager = "apk" // PackageManagerApk represents the apk package manager.
	PackageManagerApt AptPackageManager = "apt" // PackageManagerApt represents the apt package manager.
)

// Validate validates the installer configuration.
func (i *AptInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	return errors
}

// runRepoUpdate runs the package manager's repo update according to the configured mode.
func (i *AptInstaller) runRepoUpdate() error {
	mode := i.Config.GetRepoUpdateMode(i.Info.Type)
	switch mode {
	case appconfig.RepoUpdateNever:
		return nil
	case appconfig.RepoUpdateAlways:
		return i.RunCmdPassThrough(string(i.PackageManager), "update")
	default: // once
		return RunRepoUpdateOnce(string(i.PackageManager)+"-update", func() error {
			return i.RunCmdPassThrough(string(i.PackageManager), "update")
		})
	}
}

// Install implements IInstaller.
func (i *AptInstaller) Install() error {
	opts := i.GetOpts()
	err := i.runRepoUpdate()
	if err != nil {
		return err
	}
	args := []string{i.installVerb()}
	if i.IsVerbose() {
		if i.PackageManager == PackageManagerApk {
			args = append(args, "--verbose")
		}
	}
	if confirm := i.getConfirmArg(); confirm != "" {
		args = append(args, confirm)
	}
	if opts.InstallFlags != nil {
		args = append(args, strings.Fields(*opts.InstallFlags)...)
	} else if opts.Flags != nil {
		args = append(args, strings.Fields(*opts.Flags)...)
	}
	args = append(args, i.GetPackageSpec())
	return i.RunCmdPassThrough(string(i.PackageManager), args...)
}

// installVerb returns the package manager's install subcommand.
func (i *AptInstaller) installVerb() string {
	if i.PackageManager == PackageManagerApk {
		return "add"
	}
	return "install"
}

// getConfirmArg returns the appropriate confirmation argument for the package manager.
// For apt, it returns "-y". For apk, it returns an empty string.
func (i *AptInstaller) getConfirmArg() string {
	confirm := "-y"
	if i.PackageManager == PackageManagerApk {
		confirm = ""
	}
	return confirm
}

// Update implements IInstaller.
func (i *AptInstaller) Update() error {
	opts := i.GetOpts()
	// `upgrade` always moves to the newest candidate, so a pinned package is reinstalled at
	// its pinned version instead.
	verb := "upgrade"
	if i.GetPinnedVersion() != "" {
		verb = i.installVerb()
	}
	args := []string{verb}
	if i.IsVerbose() {
		if i.PackageManager == PackageManagerApk {
			args = append(args, "--verbose")
		}
	}
	if confirm := i.getConfirmArg(); confirm != "" {
		args = append(args, confirm)
	}
	if opts.UpdateFlags != nil {
		args = append(args, strings.Fields(*opts.UpdateFlags)...)
	} else if opts.Flags != nil {
		args = append(args, strings.Fields(*opts.Flags)...)
	}
	args = append(args, i.GetPackageSpec())
	return i.RunCmdPassThrough(string(i.PackageManager), args...)
}

// CheckNeedsUpdate implements IInstaller.
func (i *AptInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	if pinned := i.GetPinnedVersion(); pinned != "" {
		return PinnedVersionNeedsUpdate(*i.Info.Name, pinned), nil
	}
	err := i.runRepoUpdate()
	if err != nil {
		return false, err
	}
	success, err := i.RunCmdGetSuccess(string(i.PackageManager), "--simulate", "upgrade", *i.Info.Name)
	if err != nil {
		return false, err
	}
	return !success, nil
}

// CheckIsInstalled implements IInstaller.
func (i *AptInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return i.RunCmdGetSuccess(utils.GetShellWhich(), i.GetBinName())
}

// GetData implements IInstaller.
func (i *AptInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

// GetOpts returns the parsed options for the AptInstaller.
func (i *AptInstaller) GetOpts() *AptOpts {
	opts := &AptOpts{}
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

// GetPinnedVersion implements IVersionPinned. The version may come from opts.version or from
// a `=version` suffix on the package name.
func (i *AptInstaller) GetPinnedVersion() string {
	if _, version, found := strings.Cut(*i.Info.Name, "="); found {
		return version
	}
	if version := i.GetOpts().Version; version != nil {
		return *version
	}
	return ""
}

// GetPackageSpec returns the package argument passed to the package manager, as
// `<name>=<version>` when a version is pinned.
func (i *AptInstaller) GetPackageSpec() string {
	name := *i.Info.Name
	if strings.Contains(name, "=") {
		return name
	}
	if version := i.GetPinnedVersion(); version != "" {
		return name + "=" + version
	}
	return name
}

// GetBinName returns the binary name for the installer.
// It uses the BinName from the installer data if provided, otherwise it uses the installer
// name with any `=version` suffix stripped.
func (i *AptInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	name, _, _ := strings.Cut(*info.Name, "=")
	return name
}

// NewAptInstaller creates a new AptInstaller.
func NewAptInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *AptInstaller {
	var packageManager AptPackageManager
	switch installer.Type {
	case appconfig.InstallerTypeApt:
		packageManager = PackageManagerApt
	case appconfig.InstallerTypeApk:
		packageManager = PackageManagerApk
	}
	i := &AptInstaller{
		InstallerBase:  InstallerBase{Data: installer},
		Config:         cfg,
		Info:           installer,
		PackageManager: packageManager,
	}

	return i
}
