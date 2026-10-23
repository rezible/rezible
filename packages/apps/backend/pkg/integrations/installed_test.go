package integrations

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/test"
)

type InstalledSuite struct {
	test.Suite
}

func TestInstalledSuite(t *testing.T) {
	suite.Run(t, &InstalledSuite{Suite: test.NewSuite()})
}

func (s *InstalledSuite) TestMeasureRead() {
	spans := tracetest.NewSpanRecorder()
	previousProvider := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans)))
	s.T().Cleanup(func() { otel.SetTracerProvider(previousProvider) })
	metrics := test.RecordMetrics(s.T())

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

	s.Run("a successful read records its duration and size on a span", func() {
		operation := attribute.String("operation", "succeeds")
		readErr := MeasureRead(s.T().Context(), "test", "succeeds", func(context.Context) (int, error) {
			return 7, nil
		})

		s.Require().NoError(readErr)
		s.EqualValues(1, metrics.Count("rezible.backend.provider.read.duration", provider, operation, attribute.String("outcome", "succeeded")))
		s.EqualValues(1, metrics.Count("rezible.backend.provider.read.result_size", provider, operation))
		s.EqualValues(7, metrics.Sum("rezible.backend.provider.read.result_size", provider, operation))
		span := readSpan("succeeds")
		s.Require().NotNil(span)
		s.Subset(span.Attributes(), []attribute.KeyValue{provider, operation})
		s.Equal(codes.Unset, span.Status().Code)
	})

	s.Run("a failed read returns its error and records no size", func() {
		operation := attribute.String("operation", "fails")
		providerErr := errors.New("provider unavailable")
		readErr := MeasureRead(s.T().Context(), "test", "fails", func(context.Context) (int, error) {
			return 0, providerErr
		})

		s.Same(providerErr, readErr)
		s.EqualValues(1, metrics.Count("rezible.backend.provider.read.duration", provider, operation, attribute.String("outcome", "failed")))
		s.Zero(metrics.Count("rezible.backend.provider.read.result_size", provider, operation))
		span := readSpan("fails")
		s.Require().NotNil(span)
		s.Equal(codes.Error, span.Status().Code)
	})

	s.Run("reads count towards a turn only inside its tally", func() {
		ctx, totals := rezai.WithTurnUsage(s.T().Context())
		succeed := func(context.Context) (int, error) { return 1, nil }
		fail := func(context.Context) (int, error) { return 0, errors.New("provider unavailable") }

		s.Require().NoError(MeasureRead(ctx, "test", "in_turn", succeed))
		s.Require().Error(MeasureRead(ctx, "test", "in_turn", fail))
		s.Require().NoError(MeasureRead(s.T().Context(), "test", "outside_turn", succeed))

		s.Equal(2, totals().Reads)
	})
}

func (s *InstalledSuite) TestReadStep() {
	end := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	start := end.Add(-time.Hour)

	s.Run("a fractional step below the minimum is raised and recorded", func() {
		step, limit := ReadStep(start, end, 11500*time.Millisecond)

		s.Equal(12*time.Second, step)
		s.Equal("step raised from 11.5s to 12s to stay within 300 points", limit)
	})

	s.Run("a fractional step above the minimum is rounded up without a limit", func() {
		step, limit := ReadStep(start, end, 12500*time.Millisecond)

		s.Equal(13*time.Second, step)
		s.Empty(limit)
	})
}
