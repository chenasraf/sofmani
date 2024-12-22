package installer

import (
	"bufio"
	"fmt"
	"os/exec"

	"github.com/chenasraf/sofmani/appconfig"
)

type BrewInstaller struct {
	Config    *appconfig.AppConfig
	Installer *appconfig.Installer
}

func (i *BrewInstaller) Install() error {
	cmd := exec.Command("brew", "install", i.Installer.Name)
	stderr, _ := cmd.StderrPipe()
	cmd.Start()
	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanRunes)
	for scanner.Scan() {
		fmt.Print(scanner.Text())
	}
	return nil
}

func NewBrewInstaller(cfg *appconfig.AppConfig, installer *appconfig.Installer) *BrewInstaller {
	return &BrewInstaller{
		Config:    cfg,
		Installer: installer,
	}
}
