package telemetry

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	rez "github.com/rezible/rezible"
)

type Config struct {
	ServiceName string        `koanf:"service_name"`
	Logging     loggingConfig `koanf:"logging"`
	Tracing     tracingConfig `koanf:"tracing"`
}

type loggingConfig struct {
	Level     string `koanf:"level"`
	Json      bool   `koanf:"json"`
	AddSource bool   `koanf:"add_source"`
	Color     bool   `koanf:"color"`
}

type tracingConfig struct {
	Enabled bool   `koanf:"enabled"`
	Level   string `koanf:"level"`
}

func loadConfig() (Config, error) {
	cfg := Config{
		ServiceName: os.Getenv("OTEL_SERVICE_NAME"),
		Logging: loggingConfig{
			Level: "info",
			Json:  true,
		},
		Tracing: tracingConfig{
			Enabled: false,
		},
	}
	if rez.Config.DebugMode() {
		cfg.Logging = loggingConfig{
			Level: "debug",
			Json:  false,
			Color: true,
		}
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = defaultServiceName
	}
	if rez.Config != nil {
		if cfgErr := rez.Config.Unmarshal("telemetry", &cfg); cfgErr != nil {
			return cfg, fmt.Errorf("telemetry config: %w", cfgErr)
		}
	}
	return cfg, nil
}

func (cfg Config) getLogLevel() slog.Level {
	switch strings.ToLower(strings.TrimSpace(cfg.Logging.Level)) {
	case "debug":
		return slog.LevelDebug
	case "", "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}
	slog.Default().Error("unknown log level, defaulting to info", "level", cfg.Logging.Level)
	return slog.LevelInfo
}
