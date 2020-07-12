package main

import (
	"github.com/elmanelman/oracle-judge/config"
	"log"
)

const defaultConfigPath = "config.yml"

func main() {
	cfg := config.Config{}

	if err := cfg.LoadFromFile(defaultConfigPath); err != nil {
		log.Fatalf("failed to load configuration: %q", err)
	}
}
