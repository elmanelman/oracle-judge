package main

import (
	"flag"
	"github.com/elmanelman/oracle-judge/config"
	"log"
)

const defaultConfigPath = "config.yml"

func main() {
	// set up command-line flags
	var configPath string
	flag.StringVar(&configPath, "cfg", defaultConfigPath, "configuration file path")
	flag.Parse()

	// load configuration
	cfg := config.Config{}
	if err := cfg.LoadFromFile(configPath); err != nil {
		log.Fatalf("failed to load configuration: %s", err)
	}
}
