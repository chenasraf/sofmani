package main

import (
	"os"

	"github.com/chenasraf/sofmani/installer"
	"github.com/chenasraf/sofmani/logger"
)

func main() {
	logFile := "sofmani.log"
	logger.InitLogger(logFile)
	cfg, err := LoadConfig()
	if err != nil {
		logger.Error("Error loading config: %v", err)
		return
	}

	logger.Info("Installing...")
	for _, i := range cfg.Install {
		logger.Info("Installing %s", i.Name)
		err, installerInstance := installer.GetInstaller(cfg, &i)
		if err != nil {
			logger.Error("%s", err)
			return
		}
		err = installer.RunInstaller(cfg, installerInstance)
		if err != nil {
			logger.Error("%s", err)
			os.Exit(1)
		}
	}
}
