package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	otelattribute "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otellog "go.opentelemetry.io/otel/log"
	otelmetric "go.opentelemetry.io/otel/metric"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

type (
	KeyValue = otelattribute.KeyValue

	TracerProvider = oteltrace.TracerProvider

	LoggerProvider otellog.LoggerProvider

	MeterProvider              = otelmetric.MeterProvider
	Meter                      = otelmetric.Meter
	MeterOption                = otelmetric.MeterOption
	Int64Counter               = otelmetric.Int64Counter
	Float64Histogram           = otelmetric.Float64Histogram
	Int64ObservableGauge       = otelmetric.Int64ObservableGauge
	Observer                   = otelmetric.Observer
	Observable                 = otelmetric.Observable
	MeasurementOption          = otelmetric.MeasurementOption
	Registration               = otelmetric.Registration
	Int64CounterOption         = otelmetric.Int64CounterOption
	Float64HistogramOption     = otelmetric.Float64HistogramOption
	Int64ObservableGaugeOption = otelmetric.Int64ObservableGaugeOption
)

func initOpenTelemetry(ctx context.Context, cfg Config) (*Service, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	res, resErr := makeResource(ctx, cfg)
	if resErr != nil {
		return nil, resErr
	}

	var tp TracerProvider
	if !isTracingEnabled(cfg) {
		tp = nooptrace.NewTracerProvider()
	} else {
		traceExporter, traceExporterErr := otlptracegrpc.New(ctx)
		if traceExporterErr != nil {
			return nil, fmt.Errorf("otlp trace exporter: %w", traceExporterErr)
		}
		sdkTp := sdktrace.NewTracerProvider(sdktrace.WithResource(res), sdktrace.WithBatcher(traceExporter))
		shutdownFns = append(shutdownFns, sdkTp.Shutdown)
		tp = sdkTp
	}
	otel.SetTracerProvider(tp)

	var mp MeterProvider
	if !isMetricsEnabled(cfg) {
		mp = noopmetric.NewMeterProvider()
	} else {
		metricExporter, metricExporterErr := otlpmetricgrpc.New(ctx)
		if metricExporterErr != nil {
			return nil, fmt.Errorf("otlp metric exporter: %w", metricExporterErr)
		}
		readerOpts := []sdkmetric.PeriodicReaderOption{}
		if cfg.Metrics.Interval > 0 {
			readerOpts = append(readerOpts, sdkmetric.WithInterval(cfg.Metrics.Interval))
		}
		sdkMp := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, readerOpts...)),
		)
		shutdownFns = append(shutdownFns, sdkMp.Shutdown)
		mp = sdkMp
	}
	otel.SetMeterProvider(mp)

	//var lp LoggerProvider
	//if !isLoggingEnabled(cfg) {
	//	lp = nooplog.NewLoggerProvider()
	//}
	//otel.SetLogger(lp)

	if startErr := runtime.Start(runtime.WithMeterProvider(mp)); startErr != nil {
		slog.Warn("failed to start runtime metrics", "error", startErr)
	}

	return NewService(mp, tp), nil
}

func makeResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	resOpts := []resource.Option{
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(StringAttr("service.name", cfg.ServiceName)),
	}

	res, resErr := resource.New(ctx, resOpts...)
	if resErr != nil {
		return nil, fmt.Errorf("otel resource: %w", resErr)
	}
	return res, nil
}

func isOtelDisabled() bool {
	return strings.EqualFold(os.Getenv("OTEL_SDK_DISABLED"), "true")
}

func isTracingEnabled(cfg Config) bool {
	if isOtelDisabled() {
		return false
	}
	tracesExporter := strings.ToLower(strings.TrimSpace(os.Getenv("OTEL_TRACES_EXPORTER")))
	if tracesExporter == "none" {
		return false
	}
	if cfg.Tracing.Enabled {
		return true
	}
	if tracesExporter == "otlp" {
		return true
	}
	for _, key := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
	} {
		if os.Getenv(key) != "" {
			return true
		}
	}
	return false
}

func isMetricsEnabled(cfg Config) bool {
	if isOtelDisabled() {
		return false
	}
	metricsExporter := strings.ToLower(strings.TrimSpace(os.Getenv("OTEL_METRICS_EXPORTER")))
	if metricsExporter == "none" {
		return false
	}
	if cfg.Metrics.Enabled {
		return true
	}
	if metricsExporter == "otlp" {
		return true
	}
	for _, key := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
	} {
		if os.Getenv(key) != "" {
			return true
		}
	}
	return false
}

func isLoggingEnabled(cfg Config) bool {
	if strings.EqualFold(os.Getenv("OTEL_SDK_DISABLED"), "true") {
		return false
	}
	logsExporter := strings.ToLower(strings.TrimSpace(os.Getenv("OTEL_LOGS_EXPORTER")))
	if logsExporter == "none" {
		return false
	}
	if cfg.Logging.Enabled {
		return true
	}
	if logsExporter == "otlp" {
		return true
	}
	for _, key := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
	} {
		if os.Getenv(key) != "" {
			return true
		}
	}
	return false
}
