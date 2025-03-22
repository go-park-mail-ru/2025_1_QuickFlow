package redis

import (
	get_env "quickflow/utils/get-env"
)

const (
	backupURL string = "localhost:6379"
)

type RedisConfig struct {
	redisURL string
}

func NewRedisConfig() *RedisConfig {
	return &RedisConfig{
		redisURL: get_env.GetEnv("REDIS_URL", backupURL),
	}
}

func (p *RedisConfig) GetURL() string {
	return p.redisURL
}
