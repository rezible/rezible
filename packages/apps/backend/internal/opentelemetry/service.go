package opentelemetry

import (
	"context"
	"errors"
	"log/slog"

	rez "github.com/rezible/rezible"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultMeterName = "github.com/rezible/rezible"
)

type Service struct {
	logger         *slog.Logger
	meterProvider  metric.MeterProvider
	tracerProvider trace.TracerProvider

	shutdownFns []func(context.Context) error
}

func (s *Service) Shutdown(ctx context.Context) error {
	var err error
	for i := len(s.shutdownFns) - 1; i >= 0; i-- {
		err = errors.Join(err, s.shutdownFns[i](ctx))
	}
	return err
}

func (s *Service) NewLogger(opts rez.NewLoggerOptions) *slog.Logger {
	return NewLogger(opts)
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}

func (s *Service) MeterProvider() metric.MeterProvider {
	return s.meterProvider
}

func (s *Service) TracerProvider() trace.TracerProvider {
	return s.tracerProvider
}

func (s *Service) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return s.tracerProvider.Tracer(name, opts...)
}

func (s *Service) DefaultTracer() trace.Tracer {
	return s.tracerProvider.Tracer(defaultMeterName)
}

func (s *Service) Meter(name string, opts ...metric.MeterOption) metric.Meter {
	return s.meterProvider.Meter(name, opts...)
}

func (s *Service) DefaultMeter() metric.Meter {
	return s.Meter(defaultMeterName)
}
