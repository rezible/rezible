package telemetry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

const (
	defaultServiceName = "rezible-backend"
)

var shutdownFns []func(context.Context) error

func Init(ctx context.Context) error {
	cfg, cfgErr := loadConfig()
	if cfgErr != nil {
		return cfgErr
	}

	slog.SetDefault(slog.New(makeSlogHandler(os.Stderr, cfg)))

	if otelErr := initOpenTelemetry(ctx, cfg); otelErr != nil {
		return fmt.Errorf("otel: %w", otelErr)
	}

	return nil
}

func makeSlogHandler(w io.Writer, cfg Config) slog.Handler {
	opts := &slog.HandlerOptions{
		AddSource:   cfg.Logging.AddSource,
		Level:       cfg.getLogLevel(),
		ReplaceAttr: nil,
	}
	if cfg.Logging.Json {
		return slog.NewJSONHandler(w, opts)
	}
	if !cfg.Logging.Color {
		return slog.NewTextHandler(w, opts)
	}
	return tint.NewHandler(w, &tint.Options{
		Level:      opts.Level,
		TimeFormat: time.Kitchen,
	})
}

func Shutdown(ctx context.Context) error {
	var err error
	for i := len(shutdownFns) - 1; i >= 0; i-- {
		err = errors.Join(err, shutdownFns[i](ctx))
	}
	shutdownFns = nil
	return err
}
