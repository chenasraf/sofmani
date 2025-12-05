package main

import (
	"github.com/chenasraf/sofmani/appconfig"
)

// LoadConfig loads the application configuration.
// It parses command-line arguments and then parses the configuration file.
func LoadConfig() (*appconfig.AppConfig, error) {
	overrides := appconfig.ParseCliConfig()
	return loadConfigFromCli(overrides)
}

// loadConfigFromCli loads the application configuration from pre-parsed CLI config.
func loadConfigFromCli(overrides *appconfig.AppCliConfig) (*appconfig.AppConfig, error) {
	cfg, err := appconfig.ParseConfig(overrides)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
