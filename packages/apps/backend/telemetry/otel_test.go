package telemetry

import "testing"

func TestTracingEnablement(t *testing.T) {
	clearOTelEnv(t)

	if isTracingEnabled(Config{}) {
		t.Fatalf("tracing should be disabled without config or env")
	}

	if !isTracingEnabled(Config{Tracing: tracingConfig{Enabled: true}}) {
		t.Fatalf("tracing should be enabled by config")
	}

	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	if isTracingEnabled(Config{Tracing: tracingConfig{Enabled: true}}) {
		t.Fatalf("OTEL_TRACES_EXPORTER=none should disable tracing")
	}
}

func TestMetricsEnablement(t *testing.T) {
	clearOTelEnv(t)

	if isMetricsEnabled(Config{}) {
		t.Fatalf("metrics should be disabled without config or env")
	}

	if !isMetricsEnabled(Config{Metrics: metricsConfig{Enabled: true}}) {
		t.Fatalf("metrics should be enabled by config")
	}

	t.Setenv("OTEL_METRICS_EXPORTER", "none")
	if isMetricsEnabled(Config{Metrics: metricsConfig{Enabled: true}}) {
		t.Fatalf("OTEL_METRICS_EXPORTER=none should disable metrics")
	}
}

func TestOTLPEndpointEnablesMetricsAndTracing(t *testing.T) {
	clearOTelEnv(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")

	if !isTracingEnabled(Config{}) {
		t.Fatalf("OTEL_EXPORTER_OTLP_ENDPOINT should enable tracing")
	}
	if !isMetricsEnabled(Config{}) {
		t.Fatalf("OTEL_EXPORTER_OTLP_ENDPOINT should enable metrics")
	}
}

func clearOTelEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"OTEL_SDK_DISABLED",
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
		"OTEL_TRACES_EXPORTER",
		"OTEL_METRICS_EXPORTER",
	} {
		t.Setenv(key, "")
	}
}
