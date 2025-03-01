package main

import (
	"log"
	"quickflow/config"

	"quickflow/internal"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("failed to load QuickFlow configuration: %v", err)
	}

	if err = internal.Run(cfg); err != nil {
		log.Fatalf("failed to start QuickFlow: %v", err)
	}
}
