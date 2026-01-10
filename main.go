package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/installer"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/machine"
	"github.com/chenasraf/sofmani/utils"
)

//go:embed version.txt
var appVersion []byte // appVersion is embedded from version.txt and contains the application version.

// main is the entry point of the application.
func main() {
	appconfig.SetVersion(strings.TrimSpace(string(appVersion)))

	// Parse CLI config first to check for --log-file flag
	cliConfig := appconfig.ParseCliConfig()

	// Handle --log-file without value: show log file path and exit
	if cliConfig.ShowLogFile {
		fmt.Println(logger.GetLogFile())
		return
	}

	// Handle --machine-id: show machine ID and exit
	if cliConfig.ShowMachineID {
		fmt.Println(machine.GetMachineID())
		return
	}

	// Set custom log file if provided
	if cliConfig.LogFile != nil {
		logger.SetLogFile(*cliConfig.LogFile)
	}

	cfg, err := loadConfigFromCli(cliConfig)
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
	logger.Debug("Log directory: %s", logger.GetLogDir())
	logger.Debug("Log file: %s", logger.GetLogFile())
	if cacheDir, err := utils.GetCacheDir(); err == nil {
		logger.Debug("Cache directory: %s", cacheDir)
	}
	logger.Debug("Config:")
	for _, line := range cfg.GetConfigDesc() {
		logger.Debug("%s", line)
	}

	// Set MACHINE_ID environment variable
	machineID := machine.GetMachineID()
	logger.Debug("Setting env MACHINE_ID=%s", machineID)
	if err := os.Setenv("MACHINE_ID", machineID); err != nil {
		logger.Error("failed to set environment variable MACHINE_ID: %v", err)
		return
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
