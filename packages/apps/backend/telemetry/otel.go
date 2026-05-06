package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"

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
	Tracer         oteltrace.Tracer

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

	var slogHandlers []slog.Handler
	if cfg.Logging.Console.Enabled {
		slogHandlers = append(slogHandlers, makeSlogConsoleHandler(os.Stderr, cfg))
	}

	//var lp LoggerProvider
	if cfg.isOTelLoggingEnabled() {
		// unused
	} else {
		//lp = nooplog.NewLoggerProvider()
	}
	logger := slog.New(slog.NewMultiHandler(slogHandlers...))
	slog.SetDefault(logger)

	var tp TracerProvider
	if cfg.isOTelTracingEnabled() {
		traceExporter, traceExporterErr := otlptracegrpc.New(ctx)
		if traceExporterErr != nil {
			return nil, fmt.Errorf("otlp trace exporter: %w", traceExporterErr)
		}
		sdkTp := sdktrace.NewTracerProvider(sdktrace.WithResource(res), sdktrace.WithBatcher(traceExporter))
		shutdownFns = append(shutdownFns, sdkTp.Shutdown)
		tp = sdkTp
	} else {
		slog.Info("tracing disabled")
		tp = nooptrace.NewTracerProvider()
	}
	otel.SetTracerProvider(tp)

	var mp MeterProvider
	if cfg.isOTelMetricsEnabled() {
		metricExporter, metricExporterErr := otlpmetricgrpc.New(ctx)
		if metricExporterErr != nil {
			return nil, fmt.Errorf("otlp metric exporter: %w", metricExporterErr)
		}
		var readerOpts []sdkmetric.PeriodicReaderOption
		if cfg.Metrics.Interval > 0 {
			readerOpts = append(readerOpts, sdkmetric.WithInterval(cfg.Metrics.Interval))
		}
		sdkMp := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, readerOpts...)),
		)
		shutdownFns = append(shutdownFns, sdkMp.Shutdown)
		mp = sdkMp
	} else {
		slog.Info("metrics disabled")
		mp = noopmetric.NewMeterProvider()
	}
	otel.SetMeterProvider(mp)

	if startErr := runtime.Start(runtime.WithMeterProvider(mp)); startErr != nil {
		slog.Warn("failed to start runtime metrics", "error", startErr)
	}

	return NewService(logger, mp, tp), nil
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
