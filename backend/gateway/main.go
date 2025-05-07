package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"quickflow/config"
	"quickflow/config/cors"
	minio_config "quickflow/config/minio"
	postgres_config "quickflow/config/postgres"
	redis_config "quickflow/config/redis"
	"quickflow/config/server"
	validation_config "quickflow/config/validation"
	"quickflow/gateway/internal"
)

func resolveConfigPath(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	if _, ok := os.LookupEnv("RUNNING_IN_CONTAINER"); ok {
		return filepath.Join("/config", rel)
	}
	return filepath.Join("../deploy/config", rel)
}

func initCfg() (*config.Config, error) {

	serverConfigPath := flag.String("server-config", "feeder/config.toml", "Path to config file")
	corsConfigPath := flag.String("cors-config", "cors/config.toml", "Path to CORS config file")
	minioConfigPath := flag.String("minio-config", "minio/config.toml", "Path to Minio config file")
	validationConfig := flag.String("validation-config", "validation/config.toml", "Path to Validation config file")
	flag.Parse()

	serverCfg, err := server_config.Parse(resolveConfigPath(*serverConfigPath))
	if err != nil {
		return nil, fmt.Errorf("failed to load project server configuration: %v", err)
	}

	corsCfg, err := cors_config.ParseCORS(resolveConfigPath(*corsConfigPath))
	if err != nil {
		return nil, fmt.Errorf("failed to load project CORS configuration: %v", err)
	}

	minioCfg, err := minio_config.ParseMinio(resolveConfigPath(*minioConfigPath))
	if err != nil {
		return nil, fmt.Errorf("failed to load project minio configuration: %v", err)
	}

	validationCfg, err := validation_config.NewValidationConfig(resolveConfigPath(*validationConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to load project validation configuration: %v", err)
	}

	postgresCfg := postgres_config.NewPostgresConfig()
	redisCfg := redis_config.NewRedisConfig()

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
