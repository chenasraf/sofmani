package main

import (
	"github.com/chenasraf/sofmani/appconfig"
)

// LoadConfig loads the application configuration.
// It parses command-line arguments and then parses the configuration file.
func LoadConfig() (*appconfig.AppConfig, error) {
	overrides := appconfig.ParseCliConfig()
	cfg, err := appconfig.ParseConfig(overrides)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
