package installer

import (
	"fmt"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
	"github.com/chenasraf/sofmani/utils"
)

type IInstaller interface {
	GetData() *appconfig.InstallerData
	CheckIsInstalled() (bool, error)
	CheckNeedsUpdate() (bool, error)
	Install() error
	Update() error
	Validate() []ValidationError
}

type InstallerBase struct {
	Data *appconfig.InstallerData
}

func GetInstaller(config *appconfig.AppConfig, data *appconfig.InstallerData) (IInstaller, error) {
	data = InstallerWithDefaults(data, data.Type, config.Defaults)
	switch data.Type {
	case appconfig.InstallerTypeGroup:
		return NewGroupInstaller(config, data), nil
	case appconfig.InstallerTypeBrew:
		return NewBrewInstaller(config, data), nil
	case appconfig.InstallerTypeShell:
		return NewShellInstaller(config, data), nil
	case appconfig.InstallerTypeRsync:
		return NewRsyncInstaller(config, data), nil
	case appconfig.InstallerTypeNpm, appconfig.InstallerTypePnpm, appconfig.InstallerTypeYarn:
		return NewNpmInstaller(config, data), nil
	case appconfig.InstallerTypeApt, appconfig.InstallerTypeApk:
		return NewAptInstaller(config, data), nil
	case appconfig.InstallerTypePipx:
		return NewPipxInstaller(config, data), nil
	case appconfig.InstallerTypeGitHubRelease:
		return NewGitHubReleaseInstaller(config, data), nil
	case appconfig.InstallerTypeGit:
		return NewGitInstaller(config, data), nil
	case appconfig.InstallerTypeManifest:
		return NewManifestInstaller(config, data), nil
	}
	return nil, nil
}

func (i *InstallerBase) GetData() *appconfig.InstallerData {
	return i.Data
}

func (i *InstallerBase) BaseValidate() []ValidationError {
	errors := []ValidationError{}
	info := i.GetData()
	if info.Name == nil || len(*info.Name) == 0 {
		errors = append(errors, ValidationError{FieldName: "name", Message: "Name is required"})
	}
	return errors
}

func (i *InstallerBase) RunCustomUpdateCheck() (bool, error) {
	envShell := utils.GetOSShell(i.GetData().EnvShell)
	args := utils.GetOSShellArgs(*i.GetData().CheckHasUpdate)
	return utils.RunCmdGetSuccessPassThrough(i.Data.Environ(), envShell, args...)
}

func (i *InstallerBase) RunCustomInstallCheck() (bool, error) {
	envShell := utils.GetOSShell(i.GetData().EnvShell)
	args := utils.GetOSShellArgs(*i.GetData().CheckInstalled)
	return utils.RunCmdGetSuccessPassThrough(i.Data.Environ(), envShell, args...)
}

func (i *InstallerBase) HasCustomUpdateCheck() bool {
	return i.GetData().CheckHasUpdate != nil
}

func (i *InstallerBase) HasCustomInstallCheck() bool {
	return i.GetData().CheckInstalled != nil
}

func (i *InstallerBase) RunCmdAsFile(command string) error {
	data := i.GetData()
	return utils.RunCmdAsFile(data.Environ(), command, data.EnvShell)
}

func (i *InstallerBase) RunCmdPassThrough(command string, args ...string) error {
	data := i.GetData()
	return utils.RunCmdPassThrough(data.Environ(), command, args...)
}

func (i *InstallerBase) RunCmdGetSuccess(command string, args ...string) (bool, error) {
	data := i.GetData()
	return utils.RunCmdGetSuccess(data.Environ(), command, args...)
}

func (i *InstallerBase) RunCmdGetSuccessPassThrough(command string, args ...string) (bool, error) {
	data := i.GetData()
	return utils.RunCmdGetSuccessPassThrough(data.Environ(), command, args...)
}

func (i *InstallerBase) RunCmdGetOutput(command string, args ...string) ([]byte, error) {
	data := i.GetData()
	return utils.RunCmdGetOutput(data.Environ(), command, args...)
}

func RunInstaller(config *appconfig.AppConfig, installer IInstaller) error {
	info := installer.GetData()
	name := *info.Name
	curOS := platform.GetPlatform()
	logger.Debug("Checking if %s (%s) should run on %s", name, info.Type, curOS)
	env := config.Environ()
	if !installer.GetData().Platforms.GetShouldRunOnOS(curOS) {
		logger.Debug("%s should not run on %s, skipping", name, curOS)
		return nil
	}
	if !FilterInstaller(installer, config.Filter) {
		logger.Debug("%s is filtered, skipping", name)
		return nil
	}

	enabled, err := InstallerIsEnabled(installer)

	if err != nil {
		return fmt.Errorf("Failed to check if %s is enabled: %s", name, err)
	}

	if !enabled {
		logger.Debug("%s is disabled, skipping", name)
		return nil
	}

	logger.Debug("Checking %s: %s", info.Type, name)
	installed, err := installer.CheckIsInstalled()
	if err != nil {
		return err
	}
	if installed {
		logger.Debug("%s is already installed", name)

		if *config.CheckUpdates {
			logger.Info("Checking updates for %s", name)
			needsUpdate, err := installer.CheckNeedsUpdate()
			if err != nil {
				return err
			}
			if needsUpdate {
				logger.Info("Updating %s", name)
				if info.PreUpdate != nil {
					logger.Debug("Running pre-update command for %s", name)
					err := utils.RunCmdPassThrough(env, utils.GetOSShell(installer.GetData().EnvShell), utils.GetOSShellArgs(*info.PreUpdate)...)
					if err != nil {
						return err
					}
				}
				logger.Debug("Running update command for %s", name)
				installer.Update()
				if info.PostUpdate != nil {
					logger.Debug("Running post-update command for %s", name)
					err := utils.RunCmdPassThrough(env, utils.GetOSShell(installer.GetData().EnvShell), utils.GetOSShellArgs(*info.PostUpdate)...)
					if err != nil {
						return err
					}
				}
			} else {
				logger.Info("%s (%s) is up-to-date", name, info.Type)
			}
			return nil
		} else {
			return nil
		}
	} else {
		logger.Info("Installing %s: %s", installer.GetData().Type, name)
		if info.PreInstall != nil {
			logger.Debug("Running pre-install command for %s (%s)", name, info.Type)
			err := utils.RunCmdPassThrough(env, utils.GetOSShell(installer.GetData().EnvShell), utils.GetOSShellArgs(*info.PreInstall)...)
			if err != nil {
				return err
			}
		}
		logger.Debug("Running installer for %s (%s)", name, info.Type)
		err = installer.Install()
		if info.PostInstall != nil {
			logger.Debug("Running post-install command for %s (%s)", name, info.Type)
			err := utils.RunCmdPassThrough(env, utils.GetOSShell(installer.GetData().EnvShell), utils.GetOSShellArgs(*info.PostInstall)...)
			if err != nil {
				return err
			}
		}
	}
	if err != nil {
		return err
	}
	return nil
}
