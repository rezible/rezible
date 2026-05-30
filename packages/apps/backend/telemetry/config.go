package telemetry

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	rez "github.com/rezible/rezible"
)

type Config struct {
	ServiceName string        `cfg:"service_name"`
	Logging     loggingConfig `cfg:"logging"`
	Tracing     tracingConfig `cfg:"tracing"`
	Metrics     metricsConfig `cfg:"metrics"`
}

type loggingConfig struct {
	Console   loggingConsoleConfig `cfg:"console"`
	OTel      loggingOtelConfig    `cfg:"otel"`
	AddSource bool                 `cfg:"add_source"`
}

type loggingConsoleConfig struct {
	Enabled bool   `cfg:"enabled"`
	Level   string `cfg:"level"`
	Json    bool   `cfg:"json"`
	Color   bool   `cfg:"color"`
}

type loggingOtelConfig struct {
	Enabled bool `cfg:"enabled"`
}

type tracingConfig struct {
	Enabled bool   `cfg:"enabled"`
	Level   string `cfg:"level"`
}

type metricsConfig struct {
	Enabled  bool          `cfg:"enabled"`
	Interval time.Duration `cfg:"interval"`
}

func loadConfig(cl rez.ConfigLoader) (Config, error) {
	cfg := Config{
		ServiceName: os.Getenv("OTEL_SERVICE_NAME"),
		Logging: loggingConfig{
			Console: loggingConsoleConfig{
				Enabled: true,
				Level:   "info",
				Json:    true,
			},
		},
		Tracing: tracingConfig{
			Enabled: true,
		},
		Metrics: metricsConfig{
			Enabled:  true,
			Interval: 30 * time.Second,
		},
	}
	if cl.DebugMode() {
		cfg.Logging.Console = loggingConsoleConfig{
			Enabled: true,
			Level:   "debug",
			Json:    false,
			Color:   true,
		}
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = defaultServiceName
	}
	if cfgErr := cl.Unmarshal("telemetry", &cfg); cfgErr != nil {
		return cfg, fmt.Errorf("telemetry config: %w", cfgErr)
	}
	return cfg, nil
}

func (cfg Config) getSlogLogLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "", "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}
	slog.Error("unknown log level, defaulting to info", "level", raw)
	return slog.LevelInfo
}

func (cfg Config) isOtelDisabled() bool {
	return strings.EqualFold(os.Getenv("OTEL_SDK_DISABLED"), "true")
}

func (cfg Config) isOTelTracingEnabled() bool {
	return cfg.Tracing.Enabled && !cfg.isOtelDisabled()
}

func (cfg Config) isOTelMetricsEnabled() bool {
	return cfg.Metrics.Enabled && !cfg.isOtelDisabled()
}

func (cfg Config) isOTelLoggingEnabled() bool {
	return cfg.Logging.OTel.Enabled && !cfg.isOtelDisabled()
}
