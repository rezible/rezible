package test

import (
	"io"
	"log/slog"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/opentelemetry"
	"go.opentelemetry.io/otel/metric"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Telemetry is an exporter-free TelemetryService for tests. It never touches
// process globals, so suites using it may run in parallel.
type Telemetry struct {
	logger *slog.Logger
	tracer *sdktrace.TracerProvider
	meter  metric.MeterProvider
}

func NewTelemetry() *Telemetry {
	return &Telemetry{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		tracer: sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.NeverSample())),
		meter:  noopmetric.NewMeterProvider(),
	}
}

func (t *Telemetry) NewLogger(opts rez.NewLoggerOptions) *slog.Logger {
	if opts.Parent == nil {
		opts.Parent = t.logger
	}
	return opentelemetry.NewLogger(opts)
}

func (t *Telemetry) Logger() *slog.Logger {
	return t.logger
}

func (t *Telemetry) TracerProvider() trace.TracerProvider {
	return t.tracer
}

func (t *Telemetry) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return t.tracer.Tracer(name, opts...)
}

func (t *Telemetry) DefaultTracer() trace.Tracer {
	return t.Tracer("github.com/rezible/rezible")
}

func (t *Telemetry) MeterProvider() metric.MeterProvider {
	return t.meter
}

func (t *Telemetry) Meter(name string, opts ...metric.MeterOption) metric.Meter {
	return t.meter.Meter(name, opts...)
}

func (t *Telemetry) DefaultMeter() metric.Meter {
	return t.Meter("github.com/rezible/rezible")
}

var _ rez.TelemetryService = (*Telemetry)(nil)
