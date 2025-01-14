package main

import (
	"github.com/chenasraf/sofmani/appconfig"
)

func LoadConfig() (*appconfig.AppConfig, error) {
	overrides := appconfig.ParseCliConfig()
	cfg, err := appconfig.ParseConfig(overrides)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
