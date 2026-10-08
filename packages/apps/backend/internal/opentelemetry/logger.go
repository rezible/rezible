package opentelemetry

import (
	"context"
	"log/slog"

	"github.com/rezible/rezible/pkg/execution"
	"go.opentelemetry.io/otel/trace"
)

// contextHandler adds the trace and span IDs and the execution attributes of a record's context.
type contextHandler struct {
	base slog.Handler
}

func (h contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

func (h contextHandler) Handle(ctx context.Context, record slog.Record) error {
	var attrs []slog.Attr
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		attrs = append(attrs,
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)
	}
	for _, kv := range execution.Attrs(ctx) {
		attrs = append(attrs, slog.Any(string(kv.Key), kv.Value.AsInterface()))
	}
	if len(attrs) > 0 {
		record = record.Clone()
		record.AddAttrs(attrs...)
	}
	return h.base.Handle(ctx, record)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{base: h.base.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{base: h.base.WithGroup(name)}
}
