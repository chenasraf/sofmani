package main

import (
	"fmt"

	"github.com/chenasraf/sofmani/installer"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Println(fmt.Errorf("Error loading config: %v", err))
		return
	}

	for _, i := range cfg.Install {
		installer.RunInstaller(installer.NewBrewInstaller(cfg, i))
	}
}
