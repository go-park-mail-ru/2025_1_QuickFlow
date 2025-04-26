package config

import (
	"quickflow/monolith/config/cors"
	minio_config "quickflow/monolith/config/minio"
	postgresCfg "quickflow/monolith/config/postgres"
	"quickflow/monolith/config/redis"
	"quickflow/monolith/config/server"
	"quickflow/monolith/config/validation"
)

type Config struct {
	PostgresConfig   *postgresCfg.PostgresConfig
	MinioConfig      *minio_config.MinioConfig
	CORSConfig       *cors_config.CORSConfig
	RedisConfig      *redis_config.RedisConfig
	ServerConfig     *server_config.ServerConfig
	ValidationConfig *validation_config.ValidationConfig
}
