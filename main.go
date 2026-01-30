package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/cmd"
	"github.com/chenasraf/sofmani/installer"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/machine"
	"github.com/chenasraf/sofmani/summary"
	"github.com/chenasraf/sofmani/utils"
)

//go:embed version.txt
var appVersion []byte // appVersion is embedded from version.txt and contains the application version.

func init() {
	cmd.RunMain = runMain
}

// main is the entry point of the application.
func main() {
	cmd.SetVersion(strings.TrimSpace(string(appVersion)))
	cmd.Execute()
}

// runMain runs the main application logic with the given CLI config.
func runMain(cliConfig *appconfig.AppCliConfig) {
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

	// First pass: validate all installers (skip category entries)
	type installItem struct {
		installer  installer.IInstaller
		isCategory bool
		data       *appconfig.InstallerData
	}
	items := []installItem{}
	hasValidationErrors := false

	for idx := range cfg.Install {
		i := &cfg.Install[idx]

		// Handle category entries specially - they don't need validation
		if i.IsCategory() {
			items = append(items, installItem{isCategory: true, data: i})
			continue
		}

		installerInstance, err := installer.GetInstaller(cfg, i)
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
					logger.Error("%s", e.Error())
				}
			} else {
				items = append(items, installItem{installer: installerInstance, data: i})
			}
		}
	}

	if hasValidationErrors {
		logger.Error("Validation errors found, exiting. Please fix the errors and try again.")
		os.Exit(1)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	interrupted := false

	installSummary := summary.NewSummary()
	for _, item := range items {
		// Check for interrupt before each item
		select {
		case <-sigChan:
			interrupted = true
			logger.Warn("Interrupted by user")
		default:
		}
		if interrupted {
			break
		}

		// Handle category entries - just log the header
		if item.isCategory {
			logger.Category(*item.data.Category, item.data.Desc)
			continue
		}

		result, err := installer.RunInstaller(cfg, item.installer)
		if err != nil {
			logger.Error("%s", err)
			break
		}
		if result != nil {
			installSummary.Add(*result)
		}
	}

	// Print summary if enabled (default: true)
	showSummary := cfg.Summary == nil || *cfg.Summary
	if showSummary {
		installSummary.Print()
	}

	if interrupted {
		logger.Info("Cancelled")
		os.Exit(130) // Standard exit code for SIGINT
	}
	logger.Info("Complete")
}
