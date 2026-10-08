package execution

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func recordSpans(t *testing.T) *tracetest.SpanRecorder {
	recorder := tracetest.NewSpanRecorder()
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
	t.Cleanup(func() { otel.SetTracerProvider(previous) })
	return recorder
}

func userContext(t *testing.T) (context.Context, uuid.UUID) {
	userID := uuid.New()
	ctx := SetContext(t.Context(), Context{
		ActorKind:  KindUser,
		Auth:       Auth{TenantID: new(7), UserID: &userID},
		Provenance: Provenance{ID: "request-1", Source: SourceHTTP},
	})
	return ctx, userID
}

func TestDoStartsSpanWithAttributes(t *testing.T) {
	recorder := recordSpans(t)
	ctx, userID := userContext(t)
	failure := errors.New("turn failed")

	var inner []attribute.KeyValue
	doErr := Do(ctx, "agent.turn", func(ctx context.Context) error {
		inner = Attrs(ctx)
		return failure
	}, attribute.String("situation_id", "s-1"))

	require.Same(t, failure, doErr)
	want := []attribute.KeyValue{
		attribute.Int("tenant_id", 7),
		attribute.String("actor", "user"),
		attribute.String("user_id", userID.String()),
		attribute.String("request_id", "request-1"),
		attribute.String("situation_id", "s-1"),
	}
	require.Equal(t, want, inner)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, "agent.turn", spans[0].Name())
	require.Equal(t, want, spans[0].Attributes())
	require.Equal(t, codes.Error, spans[0].Status().Code)
	require.Len(t, spans[0].Events(), 1, "the error is recorded")
}

func TestDoWithoutErrorLeavesSpanStatusUnset(t *testing.T) {
	recorder := recordSpans(t)
	require.NoError(t, Do(t.Context(), "provider.read", func(context.Context) error { return nil }))
	require.Equal(t, codes.Unset, recorder.Ended()[0].Status().Code)
}

func TestWithAttrsLaterValueWins(t *testing.T) {
	ctx, _ := userContext(t)
	ctx = WithAttrs(ctx, attribute.String("situation_id", "s-1"), attribute.String("agent_turn_id", "t-1"))
	ctx = WithAttrs(ctx, attribute.String("situation_id", "s-2"), attribute.String("tenant_id", "override"))

	attrs := Attrs(ctx)
	require.Equal(t, attribute.String("tenant_id", "override"), attrs[0])
	require.Equal(t, []attribute.KeyValue{
		attribute.String("situation_id", "s-2"),
		attribute.String("agent_turn_id", "t-1"),
	}, attrs[4:])
}

func TestAnonymousContextHasNoIdentity(t *testing.T) {
	require.Empty(t, Attrs(t.Context()))

	root := NewRootContext(t.Context(), KindAnonymous, SourceHTTP)
	require.Equal(t,
		[]attribute.KeyValue{attribute.String("request_id", GetContext(root).Provenance.ID)},
		Attrs(root),
	)
	require.Equal(t,
		[]attribute.KeyValue{attribute.String("situation_id", "s-1")},
		Attrs(WithAttrs(t.Context(), attribute.String("situation_id", "s-1"))),
	)
}
