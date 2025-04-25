package config

import (
	cors_config "quickflow/config/cors"
	minio_config "quickflow/config/minio"
	"quickflow/config/postgres"
	redis_config "quickflow/config/redis"
	server_config "quickflow/config/server"
	validation_config "quickflow/config/validation"
)

type Config struct {
	PostgresConfig   *postgres_config.PostgresConfig
	MinioConfig      *minio_config.MinioConfig
	CORSConfig       *cors_config.CORSConfig
	RedisConfig      *redis_config.RedisConfig
	ServerConfig     *server_config.ServerConfig
	ValidationConfig *validation_config.ValidationConfig
}
