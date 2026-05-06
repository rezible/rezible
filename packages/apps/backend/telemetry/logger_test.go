package telemetry

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestLoggerInheritsParentFromContext(t *testing.T) {
	ctx := t.Context()
	handler := newTestHandler()
	parent := slog.New(handler).With("request_id", "req-1")
	ctx = ContextWithLogger(ctx, parent)

	logger := NewLogger(ctx, WithLogPackage("db"), WithLogAttrs(slog.String("tenant", "acme")))

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	if err := logger.Handler().Handle(ctx, record); err != nil {
		t.Fatalf("handle record: %v", err)
	}

	if !handler.hasAttr("request_id", "req-1") {
		t.Fatal("expected child logger to inherit parent attributes")
	}
	if !handler.hasAttr("package", "db") {
		t.Fatal("expected child logger to include option attributes")
	}
	if !handler.hasAttr("tenant", "acme") {
		t.Fatal("expected child logger to include slog attrs")
	}
}

func TestLoggerMinLevel(t *testing.T) {
	ctx := t.Context()
	logger := NewLogger(ctx, WithParentLogger(slog.New(newTestHandler())), WithMinLogLevel(slog.LevelWarn))

	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Fatal("expected info to be disabled")
	}
	if !logger.Enabled(ctx, slog.LevelWarn) {
		t.Fatal("expected warn to be enabled")
	}
}

type testHandler struct {
	attrs *[]slog.Attr
}

func newTestHandler() testHandler {
	attrs := []slog.Attr{}
	return testHandler{attrs: &attrs}
}

func (h testHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h testHandler) Handle(context.Context, slog.Record) error {
	return nil
}

func (h testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	*h.attrs = append(*h.attrs, attrs...)
	return h
}

func (h testHandler) WithGroup(string) slog.Handler {
	return h
}

func (h testHandler) hasAttr(key, value string) bool {
	for _, attr := range *h.attrs {
		if attr.Key == key && attr.Value.String() == value {
			return true
		}
	}
	return false
}
