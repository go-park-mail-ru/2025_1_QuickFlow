package config

import (
	"os"
)

const (
	defaultDataBaseURL string = "postgresql://user:password@localhost:5432/quickflow_db"
)

type PostgresConfig struct {
	dataBaseURL string
}

func NewPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		dataBaseURL: getEnvWithDefault("DATABASE_URL", defaultDataBaseURL),
	}
}

func (p *PostgresConfig) GetURL() string {
	return p.dataBaseURL
}

func getEnvWithDefault(name string, defaultVal string) string {
	if value, exists := os.LookupEnv(name); exists {
		return value
	}

	return defaultVal
}
