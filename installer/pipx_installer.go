package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type PipxInstaller struct {
	InstallerBase
	Config *appconfig.AppConfig
	Info   *appconfig.InstallerData
}

type PipxOpts struct {
	//
}

func (i *PipxInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	return errors
}

// Install implements IInstaller.
func (i *PipxInstaller) Install() error {
	name := *i.Info.Name
	return i.RunCmdPassThrough("pipx", "install", name)
}

// Update implements IInstaller.
func (i *PipxInstaller) Update() error {
	return i.RunCmdPassThrough("pipx", "upgrade", *i.Info.Name)
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

func (i *PipxInstaller) GetOpts() *PipxOpts {
	opts := &PipxOpts{}
	info := i.Info
	if info.Opts != nil {
		//
	}
	return opts
}

func (i *PipxInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

func NewPipxInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *PipxInstaller {
	i := &PipxInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Info:          installer,
	}

	return i
}
