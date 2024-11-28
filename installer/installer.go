package installer

import (
	"github.com/chenasraf/sofmani/appconfig"
)

type InstallerImpl struct {
	Config *appconfig.AppConfig
}

type IInstaller interface {
	Install() error
}

func RunInstaller(installer IInstaller) error {
	installer.Install()
	return nil
}
