package opentelemetry

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// exportSpan exports one span from the given instrumentation scope, failing with an error that quotes
// a prompt, and returns it as exported.
func exportSpan(t *testing.T, scope string, captureAiContent bool) tracetest.SpanStub {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(withAiContentPolicy(exporter, captureAiContent)),
	)
	_, span := provider.Tracer(scope).Start(t.Context(), "generate")
	span.SetAttributes(
		attribute.String("genkit:input", `{"messages":[{"text":"checkout-api logs"}]}`),
		attribute.String("genkit:init", `{"state":{"messages":[{"text":"earlier answer"}],"artifacts":[{"name":"report"}]}}`),
		attribute.String("genkit:name", "generate"),
		attribute.String("genkit:output", `{"text":"the deploy at 10:02"}`),
		attribute.String("genkit:state", "error"),
	)
	span.AddEvent("retry", trace.WithAttributes(attribute.Int("attempt", 2)))
	failure := errors.New(`model rejected prompt "checkout-api logs"`)
	span.RecordError(failure)
	span.SetStatus(codes.Error, failure.Error())
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	return spans[0]
}

var (
	capturedAttributes = []attribute.KeyValue{
		attribute.String("genkit:input", `{"messages":[{"text":"checkout-api logs"}]}`),
		attribute.String("genkit:init", `{"state":{"messages":[{"text":"earlier answer"}],"artifacts":[{"name":"report"}]}}`),
		attribute.String("genkit:name", "generate"),
		attribute.String("genkit:output", `{"text":"the deploy at 10:02"}`),
		attribute.String("genkit:state", "error"),
	}
	capturedStatus    = sdktrace.Status{Code: codes.Error, Description: `model rejected prompt "checkout-api logs"`}
	capturedException = []attribute.KeyValue{
		attribute.String("exception.type", "*errors.errorString"),
		attribute.String("exception.message", `model rejected prompt "checkout-api logs"`),
	}
)

func requireEvents(t *testing.T, span tracetest.SpanStub, exception []attribute.KeyValue) {
	t.Helper()
	require.Len(t, span.Events, 2)
	require.Equal(t, "retry", span.Events[0].Name)
	require.Equal(t, []attribute.KeyValue{attribute.Int("attempt", 2)}, span.Events[0].Attributes)
	require.Equal(t, "exception", span.Events[1].Name)
	require.Equal(t, exception, span.Events[1].Attributes)
}

func TestGenkitContentIsRedactedUnlessCaptured(t *testing.T) {
	redacted := exportSpan(t, genkitScope, false)
	require.Equal(t, []attribute.KeyValue{
		attribute.String("genkit:input", "<redacted>"),
		attribute.String("genkit:init", "<redacted>"),
		attribute.String("genkit:name", "generate"),
		attribute.String("genkit:output", "<redacted>"),
		attribute.String("genkit:state", "error"),
	}, redacted.Attributes)
	require.Equal(t, sdktrace.Status{Code: codes.Error, Description: "<redacted>"}, redacted.Status)
	requireEvents(t, redacted, []attribute.KeyValue{
		attribute.String("exception.type", "*errors.errorString"),
		attribute.String("exception.message", "<redacted>"),
	})

	captured := exportSpan(t, genkitScope, true)
	require.Equal(t, capturedAttributes, captured.Attributes)
	require.Equal(t, capturedStatus, captured.Status)
	requireEvents(t, captured, capturedException)
}

func TestOtherSpansAreNotRedacted(t *testing.T) {
	for _, captureAiContent := range []bool{false, true} {
		exported := exportSpan(t, "github.com/rezible/rezible/pkg/execution", captureAiContent)
		require.Equal(t, capturedAttributes, exported.Attributes)
		require.Equal(t, capturedStatus, exported.Status)
		requireEvents(t, exported, capturedException)
	}
}
