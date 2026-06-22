package watermill

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

type testTelemetryService struct {
	meterProvider  metric.MeterProvider
	tracerProvider trace.TracerProvider
}

func newTestTelemetryService() testTelemetryService {
	return testTelemetryService{
		meterProvider:  noopmetric.NewMeterProvider(),
		tracerProvider: nooptrace.NewTracerProvider(),
	}
}

func (s testTelemetryService) NewLogger(opts rez.NewLoggerOptions) *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func (s testTelemetryService) Logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func (s testTelemetryService) TracerProvider() trace.TracerProvider {
	return s.tracerProvider
}

func (s testTelemetryService) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return s.tracerProvider.Tracer(name, opts...)
}

func (s testTelemetryService) DefaultTracer() trace.Tracer {
	return s.Tracer("test")
}

func (s testTelemetryService) MeterProvider() metric.MeterProvider {
	return s.meterProvider
}

func (s testTelemetryService) Meter(name string, opts ...metric.MeterOption) metric.Meter {
	return s.meterProvider.Meter(name, opts...)
}

func (s testTelemetryService) DefaultMeter() metric.Meter {
	return s.Meter("test")
}

func TestMessageServicePublishEventUsesGoChannelPublisher(t *testing.T) {
	svc, err := NewMessageService(newTestTelemetryService())
	require.NoError(t, err)

	ctx := context.Background()
	require.NotPanics(t, func() {
		err = svc.PublishEvent(ctx, rez.EventOnIncidentUpdated{IncidentId: uuid.New()})
	})
	require.NoError(t, err)
}
