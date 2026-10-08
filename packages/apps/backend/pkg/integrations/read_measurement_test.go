package integrations

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/test"
)

func TestMeasureRead(t *testing.T) {
	spans := tracetest.NewSpanRecorder()
	previousProvider := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans)))
	t.Cleanup(func() { otel.SetTracerProvider(previousProvider) })
	metrics := test.RecordMetrics(t)

	provider := attribute.String("provider", "test")
	readSpan := func(operation string) sdktrace.ReadOnlySpan {
		for _, span := range spans.Ended() {
			attrs := attribute.NewSet(span.Attributes()...)
			value, _ := attrs.Value("operation")
			if span.Name() == "provider.read" && value.AsString() == operation {
				return span
			}
		}
		return nil
	}

	t.Run("a successful read records its duration and size on a span", func(t *testing.T) {
		operation := attribute.String("operation", "succeeds")
		readErr := MeasureRead(t.Context(), "test", "succeeds", func(context.Context) (int, error) {
			return 7, nil
		})

		require.NoError(t, readErr)
		require.EqualValues(t, 1, metrics.Count("rezible.backend.provider.read.duration", provider, operation, attribute.String("outcome", "succeeded")))
		require.EqualValues(t, 1, metrics.Count("rezible.backend.provider.read.result_size", provider, operation))
		require.EqualValues(t, 7, metrics.Sum("rezible.backend.provider.read.result_size", provider, operation))
		span := readSpan("succeeds")
		require.NotNil(t, span)
		require.Subset(t, span.Attributes(), []attribute.KeyValue{provider, operation})
		require.Equal(t, codes.Unset, span.Status().Code)
	})

	t.Run("a failed read returns its error and records no size", func(t *testing.T) {
		operation := attribute.String("operation", "fails")
		providerErr := errors.New("provider unavailable")
		readErr := MeasureRead(t.Context(), "test", "fails", func(context.Context) (int, error) {
			return 0, providerErr
		})

		require.Same(t, providerErr, readErr)
		require.EqualValues(t, 1, metrics.Count("rezible.backend.provider.read.duration", provider, operation, attribute.String("outcome", "failed")))
		require.Zero(t, metrics.Count("rezible.backend.provider.read.result_size", provider, operation))
		span := readSpan("fails")
		require.NotNil(t, span)
		require.Equal(t, codes.Error, span.Status().Code)
	})

	t.Run("reads count towards a turn only inside its tally", func(t *testing.T) {
		ctx, totals := rezai.WithTurnUsage(t.Context())
		succeed := func(context.Context) (int, error) { return 1, nil }
		fail := func(context.Context) (int, error) { return 0, errors.New("provider unavailable") }

		require.NoError(t, MeasureRead(ctx, "test", "in_turn", succeed))
		require.Error(t, MeasureRead(ctx, "test", "in_turn", fail))
		require.NoError(t, MeasureRead(t.Context(), "test", "outside_turn", succeed))

		require.Equal(t, 2, totals().Reads)
	})
}
