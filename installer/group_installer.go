package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/utils"
)

type GroupInstaller struct {
	Config *appconfig.AppConfig
	Info   *appconfig.Installer
}

type GroupOpts struct {
	BinName        *string
	CheckHasUpdate *string
	PreCommand     *string
	PostCommand    *string
}

// Install implements IInstaller.
func (i *GroupInstaller) Install() error {
	logger.Debug("Installing group %s", i.Info.Name)
	if i.GetOpts().PreCommand != nil {
		logger.Debug("Running pre-command for group %s", i.Info.Name)
		err := utils.RunCmdPassThrough("sh", "-c", *i.GetOpts().PreCommand)
		if err != nil {
			return err
		}
	}
	for _, step := range *i.Info.Steps {
		err, installer := GetInstaller(i.Config, &step)
		if err != nil {
			return err
		}
		if installer == nil {
			logger.Warn("Installer type %s is not supported, skipping", step.Type)
		} else {
			RunInstaller(i.Config, installer)
		}
	}
	if i.GetOpts().PostCommand != nil {
		logger.Debug("Running post-command for group %s", i.Info.Name)
		err := utils.RunCmdPassThrough("sh", "-c", *i.GetOpts().PostCommand)
		if err != nil {
			return err
		}
	}
	return nil
}

// Update implements IInstaller.
func (i *GroupInstaller) Update() error {
	return nil
}

// CheckNeedsUpdate implements IInstaller.
func (i *GroupInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetOpts().CheckHasUpdate != nil {
		return utils.RunCmdGetSuccess("sh", "-c", *i.GetOpts().CheckHasUpdate)
	}
	return nil, false
}

// CheckIsInstalled implements IInstaller.
func (i *GroupInstaller) CheckIsInstalled() (error, bool) {
	return utils.RunCmdGetSuccess("which", i.GetBinName())
}

// GetInfo implements IInstaller.
func (i *GroupInstaller) GetInfo() *appconfig.Installer {
	return i.Info
}

func (i *GroupInstaller) GetOpts() *GroupOpts {
	opts := &GroupOpts{}
	info := i.Info
	if info.Opts != nil {
		if binName, ok := (*info.Opts)["bin_name"].(string); ok {
			opts.BinName = &binName
		}
		if command, ok := (*info.Opts)["check_has_update"].(string); ok {
			opts.CheckHasUpdate = &command
		}
		if command, ok := (*info.Opts)["pre_command"].(string); ok {
			opts.PreCommand = &command
		}
		if command, ok := (*info.Opts)["post_command"].(string); ok {
			opts.PostCommand = &command
		}
	}
	return opts
}

func (i *GroupInstaller) GetBinName() string {
	opts := i.GetOpts()
	if opts.BinName != nil && len(*opts.BinName) > 0 {
		return *opts.BinName
	}
	return i.Info.Name
}

func NewGroupInstaller(cfg *appconfig.AppConfig, installer *appconfig.Installer) *GroupInstaller {
	return &GroupInstaller{
		Config: cfg,
		Info:   installer,
	}
}
