package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chenasraf/sofmani/appconfig"
)

func LoadConfig() (*appconfig.AppConfig, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("%s/%s", wd, "sofmani.pkl")
	fmt.Println(fmt.Sprintf("Loading config from path: %s", path))
	cfg, err := appconfig.LoadFromPath(context.Background(), path)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
