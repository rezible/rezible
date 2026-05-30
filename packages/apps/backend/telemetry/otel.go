package telemetry

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

	otelmetric "go.opentelemetry.io/otel/metric"
	noopmetric "go.opentelemetry.io/otel/metric/noop"

	oteltrace "go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

func newOpenTelemetryService(ctx context.Context, cfg Config) (*Service, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	resOpts := []sdkresource.Option{
		sdkresource.WithFromEnv(),
		sdkresource.WithTelemetrySDK(),
		sdkresource.WithAttributes(attribute.String("service.name", cfg.ServiceName)),
	}

	res, resErr := sdkresource.New(ctx, resOpts...)
	if resErr != nil {
		return nil, fmt.Errorf("otel resource: %w", resErr)
	}

	s := &Service{cfg: cfg}
	if loggerErr := s.initLogger(); loggerErr != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", loggerErr)
	}

	if tracerErr := s.initTracerProvider(ctx, res); tracerErr != nil {
		return nil, fmt.Errorf("failed to initialize tracer: %w", tracerErr)
	}

	if metricsErr := s.initMetricsProvider(ctx, res); metricsErr != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", metricsErr)
	}

	return s, nil
}

func (s *Service) initLogger() error {
	var slogHandlers []slog.Handler
	if s.cfg.Logging.Console.Enabled {
		slogHandlers = append(slogHandlers, s.makeSlogConsoleHandler(os.Stderr))
	}

	//var lp LoggerProvider
	if s.cfg.isOTelLoggingEnabled() {
		// unused
	} else {
		//lp = nooplog.NewLoggerProvider()
	}
	s.logger = slog.New(slog.NewMultiHandler(slogHandlers...))
	slog.SetDefault(s.logger)

	return nil
}

func (s *Service) makeSlogConsoleHandler(w io.Writer) slog.Handler {
	opts := &slog.HandlerOptions{
		AddSource:   s.cfg.Logging.AddSource,
		Level:       s.cfg.getSlogLogLevel(s.cfg.Logging.Console.Level),
		ReplaceAttr: nil,
	}
	if s.cfg.Logging.Console.Json {
		return slog.NewJSONHandler(w, opts)
	}
	if !s.cfg.Logging.Console.Color {
		return slog.NewTextHandler(w, opts)
	}
	return tint.NewHandler(w, &tint.Options{
		Level:      opts.Level,
		TimeFormat: time.Kitchen,
	})
}

func (s *Service) initTracerProvider(ctx context.Context, r *sdkresource.Resource) error {
	var tp oteltrace.TracerProvider
	if s.cfg.isOTelTracingEnabled() {
		traceExporter, traceExporterErr := otlptracegrpc.New(ctx)
		if traceExporterErr != nil {
			return fmt.Errorf("otlp trace exporter: %w", traceExporterErr)
		}
		sdkTp := sdktrace.NewTracerProvider(sdktrace.WithResource(r), sdktrace.WithBatcher(traceExporter))
		s.shutdownFns = append(s.shutdownFns, sdkTp.Shutdown)
		tp = sdkTp
	} else {
		slog.Info("tracing disabled")
		tp = nooptrace.NewTracerProvider()
	}
	otel.SetTracerProvider(tp)
	s.tracerProvider = tp
	return nil
}

func (s *Service) initMetricsProvider(ctx context.Context, r *sdkresource.Resource) error {
	var mp otelmetric.MeterProvider
	if s.cfg.isOTelMetricsEnabled() {
		metricExporter, metricExporterErr := otlpmetricgrpc.New(ctx)
		if metricExporterErr != nil {
			return fmt.Errorf("otlp metric exporter: %w", metricExporterErr)
		}
		var readerOpts []sdkmetric.PeriodicReaderOption
		if s.cfg.Metrics.Interval > 0 {
			readerOpts = append(readerOpts, sdkmetric.WithInterval(s.cfg.Metrics.Interval))
		}
		sdkMp := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(r),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, readerOpts...)),
		)
		s.shutdownFns = append(s.shutdownFns, sdkMp.Shutdown)
		mp = sdkMp
	} else {
		slog.Info("metrics disabled")
		mp = noopmetric.NewMeterProvider()
	}
	otel.SetMeterProvider(mp)
	if startErr := runtime.Start(runtime.WithMeterProvider(mp)); startErr != nil {
		slog.Warn("failed to start runtime metrics", "error", startErr)
	}
	s.meterProvider = mp
	return nil
}
