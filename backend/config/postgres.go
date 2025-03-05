package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultName     string = "quickflow_bd"
	defaultUserName string = "user"
	defaultPassword string = "password"
	defaultPort     int    = 5432
)

type PostgresConfig struct {
	dbName   string
	username string
	password string
	dbPort   int
}

func NewPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		dbName:   getEnvWithDefault("DB_NAME", defaultName),
		username: getEnvWithDefault("DB_USERNAME", defaultUserName),
		password: getEnvWithDefault("DB_PASSWORD", defaultPassword),
		dbPort:   getEnvAsInt("DB_PORT", defaultPort),
	}
}

func (p *PostgresConfig) GetURL() string {
	return fmt.Sprintf("postgres://%s:%s@localhost:%d/%s", p.username, p.password, p.dbPort, p.dbName)
}

func getEnvWithDefault(name string, defaultVal string) string {
	if value, exists := os.LookupEnv(name); exists {
		return value
	}

	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int {
	numString := os.Getenv(name)
	num, err := strconv.Atoi(numString)
	if err != nil {
		return defaultVal
	}

	return num
}
