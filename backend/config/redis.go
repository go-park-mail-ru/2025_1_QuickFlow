package config

import (
	"os"
)

const (
	backupURL string = "redis:6379"
)

type RedisConfig struct {
	redisURL string
}

func NewRedisConfig() *RedisConfig {
	return &RedisConfig{
		redisURL: getUrlWithDefault("REDIS_URL", backupURL),
	}
}

func (p *RedisConfig) GetURL() string {
	return p.redisURL
}

func getUrlWithDefault(name string, defaultVal string) string {
	if value, exists := os.LookupEnv(name); exists {
		return value
	}

	return defaultVal
}
