package telemetry

import (
	"sync/atomic"

	"go.opentelemetry.io/otel"
	otelattribute "go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

const defaultMeterName = "github.com/rezible/rezible"

type Service struct {
	logger         *Logger
	meterProvider  MeterProvider
	tracerProvider TracerProvider
}

var defaultService atomic.Pointer[Service]

func NewService(logger *Logger, mp MeterProvider, tp TracerProvider) *Service {
	return &Service{
		logger:         logger,
		meterProvider:  mp,
		tracerProvider: tp,
	}
}

func Default() *Service {
	if s := defaultService.Load(); s != nil {
		return s
	}
	panic("no default telemetry service initialized")
	//s := NewService(slog.Default(), otel.GetMeterProvider(), otel.GetTracerProvider())
	//if defaultService.CompareAndSwap(nil, s) {
	//	return s
	//}
	//return defaultService.Load()
}

func DefaultMeter() Meter {
	return Default().DefaultMeter()
}

func DefaultTracer() Tracer {
	return Default().DefaultTracer()
}

func (s *Service) MeterProvider() MeterProvider {
	return s.meterProvider
}

func (s *Service) TracerProvider() TracerProvider {
	return s.tracerProvider
}

func (s *Service) DefaultTracer() Tracer {
	return s.tracerProvider.Tracer(defaultMeterName)
}

func (s *Service) Meter(name string, opts ...MeterOption) Meter {
	if s == nil || s.meterProvider == nil {
		return otel.Meter(name, opts...)
	}
	return s.meterProvider.Meter(name, opts...)
}

func (s *Service) DefaultMeter() Meter {
	return s.Meter(defaultMeterName)
}

func Int64CounterInstrument(meter Meter, name, description string) Int64Counter {
	inst, err := meter.Int64Counter(name, otelmetric.WithDescription(description))
	if err != nil {
		panic(err)
	}
	return inst
}

func Float64HistogramInstrument(meter Meter, name, description, unit string) Float64Histogram {
	inst, err := meter.Float64Histogram(name, otelmetric.WithDescription(description), otelmetric.WithUnit(unit))
	if err != nil {
		panic(err)
	}
	return inst
}

func WithMetricAttributes(attributes ...KeyValue) MeasurementOption {
	return otelmetric.WithAttributes(attributes...)
}

func NormalizeLabel(value string) string {
	if value == "" {
		return "unknown"
	}
	return value
}

func StringAttr(key, value string) KeyValue {
	return otelattribute.String(key, value)
}

func BoolAttr(key string, value bool) KeyValue {
	return otelattribute.Bool(key, value)
}

func IntAttr(key string, value int) KeyValue {
	return otelattribute.Int(key, value)
}

func ResultAttr(err error) KeyValue {
	if err != nil {
		return StringAttr("result", "error")
	}
	return StringAttr("result", "success")
}
