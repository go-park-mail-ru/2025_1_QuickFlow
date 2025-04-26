package main

import (
	"flag"
	"fmt"
	"log"
	"quickflow/monolith/config"
	"quickflow/monolith/config/cors"
	minio_config "quickflow/monolith/config/minio"
	"quickflow/monolith/config/postgres"
	"quickflow/monolith/config/redis"
	"quickflow/monolith/config/server"
	"quickflow/monolith/config/validation"
	"quickflow/monolith/internal"
)

func initCfg() (*config.Config, error) {
	serverConfigPath := flag.String("server-config", "", "Path to config file")
	corsConfigPath := flag.String("cors-config", "", "Path to CORS config file")
	minioConfigPath := flag.String("minio-config", "", "Path to Minio config file")
	validationConfig := flag.String("validation-config", "", "Path to Validation config file")
	flag.Parse()

	serverCfg, err := server_config.server_config.Parse(*serverConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load project server configuration: %v", err)
	}

	postgresCfg := postgres_config.NewPostgresConfig()
	redisCfg := redis_config.NewRedisConfig()

	corsCfg, err := cors_config.cors_config.ParseCORS(*corsConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load project CORS configuration: %v", err)
	}

	minioCfg, err := minio_config.ParseMinio(*minioConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load project minio configuration: %v", err)
	}

	validationCfg, err := validation_config.NewValidationConfig(*validationConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load project validation configuration: %v", err)
	}

	return &config.Config{
		PostgresConfig:   postgresCfg,
		ServerConfig:     serverCfg,
		CORSConfig:       corsCfg,
		MinioConfig:      minioCfg,
		RedisConfig:      redisCfg,
		ValidationConfig: validationCfg,
	}, nil
}

func main() {
	appCfg, err := initCfg()
	if err != nil {
		log.Fatalf("failed to initialize configuration: %v", err)
	}

	if err = internal.Run(appCfg); err != nil {
		log.Fatalf("failed to start QuickFlow: %v", err)
	}
}
