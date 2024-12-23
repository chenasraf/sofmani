package main

import (
	"fmt"
	"os"

	"github.com/chenasraf/sofmani/installer"
	"github.com/chenasraf/sofmani/logger"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Println(fmt.Errorf("Error loading config: %v", err))
		return
	}
	logFile := "sofmani.log"
	logger.InitLogger(logFile, cfg)

	logger.Info("Installing...")
	for _, i := range cfg.Install {
		err, installerInstance := installer.GetInstaller(cfg, &i)
		if err != nil {
			logger.Error("%s", err)
			return
		}
		if installerInstance == nil {
			logger.Warn("Installer type %s is not supported, skipping", i.Type)
		} else {
			err = installer.RunInstaller(cfg, installerInstance)
			if err != nil {
				logger.Error("%s", err)
				os.Exit(1)
			}
		}
	}
}
