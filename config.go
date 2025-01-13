package main

import (
	"github.com/chenasraf/sofmani/appconfig"
)

func LoadConfig(version string) (*appconfig.AppConfig, error) {
	cfg, err := appconfig.ParseConfig(version)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
