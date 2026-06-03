package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig    `yaml:"server"`
	Backends []BackendConfig `yaml:"backends"`
}

type ServerConfig struct {
	Port                int    `yaml:"port"`
	HealthCheckInterval string `yaml:"health_check_interval"`
}

type BackendConfig struct {
	URL string `yaml:"url"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed load config: %w", err)
	}

	var cfg Config

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return errors.New("invalid port")
	}
	if len(c.Backends) == 0 {
		return errors.New("empty backends")
	}

	return nil
}
