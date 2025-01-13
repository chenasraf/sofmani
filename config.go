package main

import (
	"github.com/chenasraf/sofmani/appconfig"
)

func LoadConfig(version string) (*appconfig.AppConfig, error) {
	overrides := appconfig.ParseCliConfig(version)
	cfg, err := appconfig.ParseConfig(version, overrides)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
