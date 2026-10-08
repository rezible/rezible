package opentelemetry

import (
	"context"
	"errors"
	"log/slog"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Service owns the process's logger and OpenTelemetry providers. Code reaches them through
// slog and the OpenTelemetry globals, which Init installs.
type Service struct {
	logger         *slog.Logger
	meterProvider  metric.MeterProvider
	tracerProvider trace.TracerProvider

	shutdownFns []func(context.Context) error
}

// Shutdown flushes and stops the exporters.
func (s *Service) Shutdown(ctx context.Context) error {
	var err error
	for i := len(s.shutdownFns) - 1; i >= 0; i-- {
		err = errors.Join(err, s.shutdownFns[i](ctx))
	}
	return err
}
