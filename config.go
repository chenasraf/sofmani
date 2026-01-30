package main

import (
	"github.com/chenasraf/sofmani/appconfig"
)

// loadConfigFromCli loads the application configuration from pre-parsed CLI config.
func loadConfigFromCli(overrides *appconfig.AppCliConfig) (*appconfig.AppConfig, error) {
	cfg, err := appconfig.ParseConfig(overrides)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
