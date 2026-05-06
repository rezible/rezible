package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
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

	svc, otelErr := initOpenTelemetry(ctx, cfg)
	if otelErr != nil {
		return fmt.Errorf("otel: %w", otelErr)
	}
	SetDefault(svc)

	return nil
}

func Shutdown(ctx context.Context) error {
	var err error
	for i := len(shutdownFns) - 1; i >= 0; i-- {
		err = errors.Join(err, shutdownFns[i](ctx))
	}
	shutdownFns = nil
	return err
}
