package postgres_config

import (
	getenv "quickflow/monolith/utils/get-env"
)

const (
	defaultDataBaseURL string = "postgresql://quickflow_admin:SuperSecurePassword1@localhost:5432/quickflow_db"
)

type PostgresConfig struct {
	dataBaseURL string
}

func NewPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		dataBaseURL: getenv.GetEnv("DATABASE_URL", defaultDataBaseURL),
	}
}

func (p *PostgresConfig) GetURL() string {
	return p.dataBaseURL
}
