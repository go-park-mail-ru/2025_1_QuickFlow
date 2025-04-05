package main

import (
	"flag"
	"log"

	"quickflow/config"
	"quickflow/config/cors"
	minio_config "quickflow/config/minio"
	"quickflow/internal"
)

func main() {
	mainConfigPath := flag.String("config", "", "Path to config file")
	corsConfigPath := flag.String("cors-config", "", "Path to CORS config file")
	minioConfigPath := flag.String("minio-config", "", "Path to Minio config file")
	flag.Parse()
	cfg, err := config.Parse(*mainConfigPath)
	if err != nil {
		log.Fatalf("failed to load QuickFlow configuration: %v", err)
	}

	corsCfg, err := cors.ParseCORS(*corsConfigPath)
	if err != nil {
		log.Fatalf("failed to load CORS configuration: %v", err)
	}

	minioCfg, err := minio_config.ParseMinio(*minioConfigPath)
	if err != nil {
		log.Fatalf("failed to load Minio configuration: %v", err)
	}

	if err = internal.Run(cfg, corsCfg, minioCfg); err != nil {
		log.Fatalf("failed to start QuickFlow: %v", err)
	}
}
