package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/installer"
	"github.com/chenasraf/sofmani/logger"
)

//go:embed version.txt
var appVersion []byte // appVersion is embedded from version.txt and contains the application version.

// main is the entry point of the application.
func main() {
	appconfig.SetVersion(strings.TrimSpace(string(appVersion)))
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Println(fmt.Errorf("error loading config: %v", err))
		return
	}
	isDebug := false
	if cfg.Debug != nil {
		isDebug = *cfg.Debug
	}
	logger.InitLogger(isDebug)

	logger.Debug("Sofmani version %s", appconfig.AppVersion)
	logger.Debug("Config:")
	for _, line := range cfg.GetConfigDesc() {
		logger.Debug("%s", line)
	}

	if cfg.Env != nil {
		for k, v := range *cfg.Env {
			logger.Debug("Setting env %s=%s", k, v)
			err := os.Setenv(k, v)
			if err != nil {
				logger.Error("failed to set environment variable %s: %v", k, err)
				return
			}
		}
	}

	logger.Info("Checking all installers...")
	instances := []installer.IInstaller{}
	hasValidationErrors := false

	for _, i := range cfg.Install {
		installerInstance, err := installer.GetInstaller(cfg, &i)
		if err != nil {
			logger.Error("%s", err)
			return
		}
		if installerInstance == nil {
			logger.Warn("Installer type %s is not supported, skipping", i.Type)
		} else {
			errors := installerInstance.Validate()
			if len(errors) > 0 {
				hasValidationErrors = true
				for _, e := range errors {
					logger.Error(e.Error())
				}
			} else {
				instances = append(instances, installerInstance)
			}
		}
	}

	if hasValidationErrors {
		logger.Error("Validation errors found, exiting. Please fix the errors and try again.")
		os.Exit(1)
	}

	for _, i := range instances {
		err = installer.RunInstaller(cfg, i)
		if err != nil {
			logger.Error("%s", err)
			os.Exit(1)
		}
	}
	logger.Info("Complete")
}
