package telemetry

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	rez "github.com/rezible/rezible"
)

type Config struct {
	ServiceName   string `koanf:"service_name"`
	LogLevel      string `koanf:"log_level"`
	ConsoleFormat string `koanf:"console_format"`

	slogLogLevel slog.Level
}

func loadConfig() (Config, error) {
	cfg := Config{
		ServiceName:   os.Getenv("OTEL_SERVICE_NAME"),
		ConsoleFormat: "json",
		LogLevel:      "info",
	}
	if rez.Config.DebugMode() {
		cfg.LogLevel = "debug"
		cfg.ConsoleFormat = "text"
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = defaultServiceName
	}
	if rez.Config != nil {
		if cfgErr := rez.Config.Unmarshal("telemetry", &cfg); cfgErr != nil {
			return cfg, fmt.Errorf("telemetry config: %w", cfgErr)
		}
	}
	var levelErr error
	cfg.slogLogLevel, levelErr = cfg.parseLogLevel(cfg.LogLevel)
	if levelErr != nil {
		return cfg, fmt.Errorf("telemetry config: %w", levelErr)
	}
	return cfg, nil
}

func (cfg Config) parseLogLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug, nil
	case "", "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown telemetry.log_level %q", raw)
	}
}

func (cfg Config) makeSlogConsoleHandler() slog.Handler {
	opts := &slog.HandlerOptions{Level: cfg.slogLogLevel}
	if strings.ToLower(strings.TrimSpace(cfg.ConsoleFormat)) != "text" {
		return slog.NewJSONHandler(os.Stderr, opts)
	}
	return tint.NewHandler(os.Stderr, &tint.Options{
		Level:      cfg.slogLogLevel,
		TimeFormat: time.Kitchen,
	})
}
