package redis_config

import (
	get_env "quickflow/utils/get-env"
)

const (
	backupURL  string = "localhost:6379"
	backupPass string = "22848Amogus"
)

type RedisConfig struct {
	redisURL  string
	redisPass string
}

func NewRedisConfig() *RedisConfig {
	return &RedisConfig{
		redisURL:  get_env.GetEnv("REDIS_URL", backupURL),
		redisPass: get_env.GetEnv("REDIS_PASS", backupPass),
	}
}

func (p *RedisConfig) GetURL() string {
	return p.redisURL
}

func (p *RedisConfig) GetPass() string {
	return p.redisPass
}
