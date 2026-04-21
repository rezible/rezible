package telemetry

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

func initOpenTelemetry(ctx context.Context, cfg Config) error {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if !isTracingEnabled() {
		otel.SetTracerProvider(nooptrace.NewTracerProvider())
		return nil
	}

	resOpts := []resource.Option{
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(attribute.String("service.name", cfg.ServiceName)),
	}
	res, resErr := resource.New(ctx, resOpts...)
	if resErr != nil {
		return fmt.Errorf("otel resource: %w", resErr)
	}

	traceExporter, traceExporterErr := otlptracegrpc.New(ctx)
	if traceExporterErr != nil {
		return fmt.Errorf("otlp trace exporter: %w", traceExporterErr)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithResource(res), sdktrace.WithBatcher(traceExporter))
	shutdownFns = append(shutdownFns, tp.Shutdown)

	otel.SetTracerProvider(tp)

	return nil
}

func isTracingEnabled() bool {
	if strings.EqualFold(os.Getenv("OTEL_SDK_DISABLED"), "true") {
		return false
	}
	for _, key := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_TRACES_EXPORTER",
	} {
		if os.Getenv(key) != "" {
			return true
		}
	}
	return false
}
