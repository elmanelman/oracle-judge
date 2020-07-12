package main

import (
	"flag"
	"github.com/elmanelman/oracle-judge/config"
	"log"
)

const defaultConfigPath = "config.yml"

func main() {
	var configPath string
	flag.StringVar(&configPath, "cfg", defaultConfigPath, "configuration file path")
	flag.Parse()

	cfg := config.Config{}
	if err := cfg.LoadFromFile(configPath); err != nil {
		log.Fatalf("failed to load configuration: %s", err)
	}
}
