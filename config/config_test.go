package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AlexandrKudryavtsev/go-load-balancer/pkg/logger"
)

func newValidConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:            8000,
			ReadTimeout:     Duration{Duration: 5 * time.Second},
			WriteTimeout:    Duration{Duration: 5 * time.Second},
			ShutdownTimeout: Duration{Duration: 5 * time.Second},
		},
		Backends: []BackendConfig{
			{
				URL:        "http://localhost:8001",
				HealthPath: "/health",
				Weight:     5,
			},
			{
				URL:        "http://localhost:8002",
				HealthPath: "/health",
				Weight:     1,
			},
			{
				URL:        "http://localhost:8003",
				HealthPath: "/health",
				Weight:     2,
			},
		},
		HealthCheck: HealthCheckConfig{
			Interval: Duration{Duration: 10 * time.Second},
		},
		Logger: logger.Config{
			Format:    "text",
			Level:     "info",
			AddSource: false,
		},
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := newValidConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
}

func TestConfigValidateInvalid(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{
			name: "invalid port",
			mutate: func(cfg *Config) {
				cfg.Server.Port = 0
			},
		},
		{
			name: "invalid backends",
			mutate: func(cfg *Config) {
				cfg.Backends = nil
			},
		},
		{
			name: "invalid backend url",
			mutate: func(cfg *Config) {
				cfg.Backends[0].URL = "localhost:8001"
			},
		},
		{
			name: "malformed backend url",
			mutate: func(cfg *Config) {
				cfg.Backends[0].URL = "http://%zz"
			},
		},
		{
			name: "empty backend url",
			mutate: func(cfg *Config) {
				cfg.Backends[0].URL = ""
			},
		},
		{
			name: "invalid backend health path empty",
			mutate: func(cfg *Config) {
				cfg.Backends[0].HealthPath = ""
			},
		},
		{
			name: "invalid backend health path without slash",
			mutate: func(cfg *Config) {
				cfg.Backends[0].HealthPath = "health"
			},
		},
		{
			name: "invalid backend weight",
			mutate: func(cfg *Config) {
				cfg.Backends[0].Weight = 0
			},
		},
		{
			name: "invalid health check interval",
			mutate: func(cfg *Config) {
				cfg.HealthCheck.Interval = Duration{}
			},
		},
		{
			name: "invalid read timeout",
			mutate: func(cfg *Config) {
				cfg.Server.ReadTimeout = Duration{}
			},
		},
		{
			name: "invalid write timeout",
			mutate: func(cfg *Config) {
				cfg.Server.WriteTimeout = Duration{}
			},
		},
		{
			name: "invalid shutdown timeout",
			mutate: func(cfg *Config) {
				cfg.Server.ShutdownTimeout = Duration{}
			},
		},
		{
			name: "invalid logger format",
			mutate: func(cfg *Config) {
				cfg.Logger.Format = "xml"
			},
		},
		{
			name: "invalid logger level",
			mutate: func(cfg *Config) {
				cfg.Logger.Level = "trace"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newValidConfig()
			tt.mutate(cfg)

			if err := cfg.Validate(); err == nil {
				t.Errorf("Validate() expected error, got nil")
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	data := []byte(`server:
  port: 8080
  read_timeout: '5s'
  write_timeout: '5s'
  shutdown_timeout: '5s'

backends:
  - url: "http://localhost:9001"
    health_path: "/health"
    weight: 5
  - url: "http://localhost:9002"
    health_path: "/health"
    weight: 3
  - url: "http://localhost:9003"
    health_path: "/health"
    weight: 1

health_check:
  interval: '10s'

logger:
  level: debug
  format: json
  add_source: false
`)

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("WriteFile() unexpected error: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() unexpected error: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 8080)
	}

	if cfg.Server.ReadTimeout.Duration != 5*time.Second {
		t.Errorf(
			"Server.ReadTimeout.Duration = %v, want %v",
			cfg.Server.ReadTimeout.Duration, 5*time.Second,
		)
	}

	if len(cfg.Backends) != 3 {
		t.Fatalf("len(Backends) = %d, want %d", len(cfg.Backends), 3)
	}

	if cfg.Backends[0].URL != "http://localhost:9001" {
		t.Errorf("Backends[0].URL = %q, want %q", cfg.Backends[0].URL, "http://localhost:9001")
	}

	if cfg.Backends[0].HealthPath != "/health" {
		t.Errorf("Backends[0].HealthPath = %q, want %q", cfg.Backends[0].HealthPath, "/health")
	}

	if cfg.Backends[0].Weight != 5 {
		t.Errorf("Backends[0].Weight = %d, want %d", cfg.Backends[0].Weight, 5)
	}

	if cfg.HealthCheck.Interval.Duration != 10*time.Second {
		t.Errorf("HealthCheck.Interval.Duration = %v, want %v", cfg.HealthCheck.Interval.Duration, 10*time.Second)
	}

	if cfg.Logger.Level != "debug" {
		t.Errorf("Logger.Level = %q, want %q", cfg.Logger.Level, "debug")
	}

	if cfg.Logger.Format != "json" {
		t.Errorf("Logger.Format = %q, want %q", cfg.Logger.Format, "json")
	}
}

func TestLoadConfigInvalid(t *testing.T) {
	tests := []struct {
		name string
		path func(t *testing.T) string
	}{
		{
			name: "file does not exist",
			path: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing.yaml")
			},
		},
		{
			name: "invalid duration",
			path: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "invalid.yaml")

				data := []byte(`server:
  read_timeout: nope
`)

				if err := os.WriteFile(path, data, 0644); err != nil {
					t.Fatalf("WriteFile() unexpected error: %v", err)
				}

				return path
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.path(t)

			cfg, err := LoadConfig(path)
			if err == nil {
				t.Fatalf("LoadConfig() expected error, got nil")
			}
			if cfg != nil {
				t.Fatalf("LoadConfig() config = %#v, want nil", cfg)
			}
		})
	}
}
