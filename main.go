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

	fmt.Println("Installing...")
	for _, i := range cfg.Install {
		fmt.Println(fmt.Sprintf("Installing %s", i.Name))
		installer.RunInstaller(installer.NewBrewInstaller(cfg, i))
	}
}
