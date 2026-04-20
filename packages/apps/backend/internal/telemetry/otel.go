package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log"
	globallog "go.opentelemetry.io/otel/log/global"
	nooplog "go.opentelemetry.io/otel/log/noop"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

func initOpentelemetry(ctx context.Context, cfg Config) ([]slog.Handler, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	var tracerProvider trace.TracerProvider
	var loggerProvider log.LoggerProvider

	var slogHandlers []slog.Handler

	if isOtelEnabled() {
		resOpts := []resource.Option{
			resource.WithFromEnv(),
			resource.WithTelemetrySDK(),
			resource.WithAttributes(attribute.String("service.name", cfg.ServiceName)),
		}
		res, resErr := resource.New(ctx, resOpts...)
		if resErr != nil {
			return nil, fmt.Errorf("otel resource: %w", resErr)
		}

		traceExporter, traceExporterErr := otlptracegrpc.New(ctx)
		if traceExporterErr != nil {
			return nil, fmt.Errorf("otlp trace exporter: %w", traceExporterErr)
		}
		tp := sdktrace.NewTracerProvider(sdktrace.WithResource(res), sdktrace.WithBatcher(traceExporter))
		shutdownFns = append(shutdownFns, tp.Shutdown)

		logExporter, logExporterErr := otlploggrpc.New(ctx)
		if logExporterErr != nil {
			return nil, fmt.Errorf("otlp log exporter: %w", logExporterErr)
		}
		lp := sdklog.NewLoggerProvider(sdklog.WithResource(res), sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)))
		shutdownFns = append(shutdownFns, lp.Shutdown)
		slogHandlers = append(slogHandlers, otelslog.NewHandler("rezible", otelslog.WithLoggerProvider(lp)))

		tracerProvider = tp
		loggerProvider = lp
	} else {
		tracerProvider = nooptrace.NewTracerProvider()
		loggerProvider = nooplog.NewLoggerProvider()
	}
	otel.SetTracerProvider(tracerProvider)
	globallog.SetLoggerProvider(loggerProvider)

	return slogHandlers, nil
}

func isOtelEnabled() bool {
	if strings.EqualFold(os.Getenv("OTEL_SDK_DISABLED"), "true") {
		return false
	}
	for _, key := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
		"OTEL_TRACES_EXPORTER",
		"OTEL_LOGS_EXPORTER",
	} {
		if os.Getenv(key) != "" {
			return true
		}
	}
	return false
}
