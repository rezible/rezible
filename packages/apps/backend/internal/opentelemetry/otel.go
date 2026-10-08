package opentelemetry

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	rez "github.com/rezible/rezible"
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
)

// NewOpenTelemetryService builds providers without touching process globals.
// Call Init once from the process entrypoint to install them.
func NewOpenTelemetryService(ctx context.Context, cfg rez.Config) (*Service, error) {
	s := &Service{}

	resOpts := []sdkresource.Option{
		sdkresource.WithFromEnv(),
		sdkresource.WithTelemetrySDK(),
		sdkresource.WithAttributes(attribute.String("service.name", cfg.Telemetry.ServiceName)),
	}

	res, resErr := sdkresource.New(ctx, resOpts...)
	if resErr != nil {
		return nil, fmt.Errorf("otel resource: %w", resErr)
	}

	logCfg := cfg.Telemetry.Logging
	if cfg.App.DebugMode {
		logCfg.Console = rez.LoggingConsoleConfig{
			Enabled: true,
			Level:   "debug",
			Json:    false,
			Color:   true,
		}
	}

	if loggerErr := s.initLogger(logCfg); loggerErr != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", loggerErr)
	}

	if tracerErr := s.initTracerProvider(ctx, res, cfg.Telemetry.Tracing); tracerErr != nil {
		return nil, fmt.Errorf("failed to initialize tracer: %w", tracerErr)
	}

	if metricsErr := s.initMetricsProvider(ctx, res, cfg.Telemetry.Metrics); metricsErr != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", metricsErr)
	}

	return s, nil
}

// Init installs the service as the process-wide OpenTelemetry and slog default.
// Genkit reads the global tracer provider and requires an SDK implementation.
func (s *Service) Init() error {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	otel.SetTracerProvider(s.tracerProvider)
	otel.SetMeterProvider(s.meterProvider)
	slog.SetDefault(s.logger)
	return runtime.Start(runtime.WithMeterProvider(s.meterProvider))
}

func IsOtelEnvDisabled() bool {
	return strings.EqualFold(os.Getenv("OTEL_SDK_DISABLED"), "true")
}

func (s *Service) initLogger(cfg rez.LoggingConfig) error {
	var slogHandlers []slog.Handler
	if cfg.Console.Enabled {
		slogHandlers = append(slogHandlers, contextHandler{base: s.makeSlogConsoleHandler(os.Stdout, cfg)})
	}
	s.logger = slog.New(slog.NewMultiHandler(slogHandlers...))

	return nil
}

func (s *Service) makeSlogConsoleHandler(w io.Writer, cfg rez.LoggingConfig) slog.Handler {
	opts := &slog.HandlerOptions{
		AddSource:   cfg.AddSource,
		Level:       slog.LevelInfo,
		ReplaceAttr: nil,
	}

	switch strings.ToLower(strings.TrimSpace(cfg.Console.Level)) {
	case "debug":
		opts.Level = slog.LevelDebug
	case "info":
		opts.Level = slog.LevelInfo
	case "warn":
		opts.Level = slog.LevelWarn
	case "error":
		opts.Level = slog.LevelError
	}
	if cfg.Console.Json {
		return slog.NewJSONHandler(w, opts)
	}
	if !cfg.Console.Color {
		return slog.NewTextHandler(w, opts)
	}
	return tint.NewTextHandler(w, &tint.Options{
		Level:      opts.Level,
		TimeFormat: time.Kitchen,
	})
}

func (s *Service) initTracerProvider(ctx context.Context, r *sdkresource.Resource, cfg rez.TracingConfig) error {
	tpOpts := []sdktrace.TracerProviderOption{sdktrace.WithResource(r)}
	if !IsOtelEnvDisabled() && cfg.Enabled {
		traceExporter, traceExporterErr := otlptracegrpc.New(ctx)
		if traceExporterErr != nil {
			return fmt.Errorf("otlp trace exporter: %w", traceExporterErr)
		}
		tpOpts = append(tpOpts, sdktrace.WithBatcher(withAiContentPolicy(traceExporter, cfg.CaptureAiContent)))
	} else {
		s.logger.Info("tracing disabled")
	}
	tp := sdktrace.NewTracerProvider(tpOpts...)
	s.shutdownFns = append(s.shutdownFns, tp.Shutdown)
	s.tracerProvider = tp
	return nil
}

func (s *Service) initMetricsProvider(ctx context.Context, r *sdkresource.Resource, cfg rez.MetricsConfig) error {
	var mp otelmetric.MeterProvider
	if !IsOtelEnvDisabled() && cfg.Enabled {
		metricExporter, metricExporterErr := otlpmetricgrpc.New(ctx)
		if metricExporterErr != nil {
			return fmt.Errorf("otlp metric exporter: %w", metricExporterErr)
		}
		var readerOpts []sdkmetric.PeriodicReaderOption
		if cfg.Interval > 0 {
			readerOpts = append(readerOpts, sdkmetric.WithInterval(cfg.Interval))
		}
		sdkMp := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(r),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, readerOpts...)),
		)
		s.shutdownFns = append(s.shutdownFns, sdkMp.Shutdown)
		mp = sdkMp
	} else {
		s.logger.Info("metrics disabled")
		mp = noopmetric.NewMeterProvider()
	}
	s.meterProvider = mp
	return nil
}
