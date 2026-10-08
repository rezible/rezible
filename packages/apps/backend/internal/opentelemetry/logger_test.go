package opentelemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/rezible/rezible/pkg/execution"
)

func newJSONContextLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(contextHandler{base: slog.NewJSONHandler(&buf, nil)}), &buf
}

func decodeLogLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line), buf.String())
	delete(line, "time")
	return line
}

func TestContextHandlerAddsTraceAndExecutionAttributes(t *testing.T) {
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider())
	t.Cleanup(func() { otel.SetTracerProvider(previous) })

	userID := uuid.New()
	ctx := execution.SetContext(t.Context(), execution.Context{
		ActorKind:  execution.KindUser,
		Auth:       execution.Auth{TenantID: new(7), UserID: &userID},
		Provenance: execution.Provenance{ID: "request-1", Source: execution.SourceHTTP},
	})
	logger, buf := newJSONContextLogger()

	var spanCtx trace.SpanContext
	doErr := execution.Do(ctx, "agent.turn", func(ctx context.Context) error {
		spanCtx = trace.SpanContextFromContext(ctx)
		logger.InfoContext(ctx, "turn started")
		return nil
	}, attribute.String("situation_id", "s-1"))
	require.NoError(t, doErr)

	require.Equal(t, map[string]any{
		"level":        "INFO",
		"msg":          "turn started",
		"trace_id":     spanCtx.TraceID().String(),
		"span_id":      spanCtx.SpanID().String(),
		"tenant_id":    float64(7),
		"actor":        "user",
		"user_id":      userID.String(),
		"request_id":   "request-1",
		"situation_id": "s-1",
	}, decodeLogLine(t, buf))
}

func TestContextHandlerLeavesRecordsWithoutContextUnchanged(t *testing.T) {
	logger, buf := newJSONContextLogger()
	logger.Info("started", "component", "river")

	require.Equal(t, map[string]any{
		"level":     "INFO",
		"msg":       "started",
		"component": "river",
	}, decodeLogLine(t, buf))
}
