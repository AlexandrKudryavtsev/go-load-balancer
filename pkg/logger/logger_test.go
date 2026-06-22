package logger

import (
	"log/slog"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "json",
			cfg:  Config{Format: FormatJSON},
		},
		{
			name: "text",
			cfg:  Config{Format: FormatText},
		},
		{
			name: "unknown format defaults to text",
			cfg:  Config{Format: "unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := New(tt.cfg)

			if log == nil {
				t.Fatal("New() = nil, want logger")
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  slog.Level
	}{
		{name: "debug", level: "debug", want: slog.LevelDebug},
		{name: "info", level: "info", want: slog.LevelInfo},
		{name: "warn", level: "warn", want: slog.LevelWarn},
		{name: "warning", level: "warning", want: slog.LevelWarn},
		{name: "error", level: "error", want: slog.LevelError},
		{name: "trims spaces", level: " \n\t   debug ", want: slog.LevelDebug},
		{name: "case insensitive", level: "DeBUg", want: slog.LevelDebug},
		{name: "unknown defaults to info", level: "super", want: slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := parseLevel(tt.level)
			if level != tt.want {
				t.Errorf("parseLevel() = %q, want %q", level, tt.want)
			}
		})
	}
}
