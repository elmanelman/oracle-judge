package main

import (
	"flag"
	"fmt"
	"github.com/elmanelman/oracle-judge/config"
	"github.com/elmanelman/oracle-judge/workers"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	// create the pipeline
	p, err := workers.NewPipeline(cfg)
	if err != nil {
		log.Fatalf("failed to create pipeline: %s", err)
	}

	// set up SIGTERM handler
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
		p.Stop()
	}()

	// start the pipeline
	if err := p.Start(); err != nil {
		log.Fatalf("failed to start pipeline: %s", err)
	}
}
