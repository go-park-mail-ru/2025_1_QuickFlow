package config

import (
    "fmt"
    "time"

    "github.com/BurntSushi/toml"
)

type Config struct {
    Addr         string        `toml:"addr"`
    ReadTimeout  time.Duration `toml:"read_timeout"`
    WriteTimeout time.Duration `toml:"write_timeout"`
}

// LoadConfig loads config from file.
func LoadConfig(configPath string) (*Config, error) {
    if configPath == "" {
        configPath = "./config/config.toml"
    }

    var cfg Config
    _, err := toml.DecodeFile(configPath, &cfg)
    if err != nil {
        return nil, fmt.Errorf("config.LoadConfig: %w", err)
    }
    return &cfg, nil
}
