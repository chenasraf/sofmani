package main

import (
	"github.com/chenasraf/sofmani/appconfig"
)

func LoadConfig() (*appconfig.AppConfig, error) {
	cfg, err := appconfig.ParseConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
