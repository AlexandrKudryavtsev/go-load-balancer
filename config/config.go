package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/AlexandrKudryavtsev/go-load-balancer/pkg/logger"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Backends    []BackendConfig   `yaml:"backends"`
	HealthCheck HealthCheckConfig `yaml:"health_check"`
	Logger      logger.Config     `yaml:"logger"`
}

type ServerConfig struct {
	Port            int      `yaml:"port"`
	ReadTimeout     Duration `yaml:"read_timeout"`
	WriteTimeout    Duration `yaml:"write_timeout"`
	ShutdownTimeout Duration `yaml:"shutdown_timeout"`
}

type BackendConfig struct {
	URL        string `yaml:"url"`
	HealthPath string `yaml:"health_path"`
	Weight     int    `yaml:"weight"`
}

type HealthCheckConfig struct {
	Interval Duration `yaml:"interval"`
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

	for i, backend := range c.Backends {
		if backend.URL == "" {
			return fmt.Errorf("backend %d: invalid url", i)
		}

		u, err := url.Parse(backend.URL)
		if err != nil {
			return fmt.Errorf("backend %d: invalid url %q: %w", i, backend.URL, err)
		}
		if u.Host == "" || u.Scheme == "" {
			return fmt.Errorf("backend %d: invalid url %q", i, backend.URL)
		}

		if backend.HealthPath == "" || !strings.HasPrefix(backend.HealthPath, "/") {
			return fmt.Errorf("backend %d: invalid health path", i)
		}
		if backend.Weight < 1 {
			return fmt.Errorf("backend %d: invalid weight", i)
		}
	}

	if c.HealthCheck.Interval.Duration <= 0 {
		return errors.New("invalid interval health check")
	}
	if c.Server.ReadTimeout.Duration <= 0 {
		return errors.New("invalid read timeout")
	}
	if c.Server.WriteTimeout.Duration <= 0 {
		return errors.New("invalid write timeout")
	}
	if c.Server.ShutdownTimeout.Duration <= 0 {
		return errors.New("invalid shutdown timeout")
	}

	if c.Logger.Format != "json" && c.Logger.Format != "text" {
		return errors.New("invalid logger format")
	}

	switch c.Logger.Level {
	case "debug", "info", "warn", "warning", "error":
	default:
		return errors.New("invalid logger level")
	}

	return nil
}
