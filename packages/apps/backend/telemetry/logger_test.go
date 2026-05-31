package telemetry

import (
	"context"
	"log/slog"
	"testing"

	rez "github.com/rezible/rezible"
)

func TestLoggerMinLevel(t *testing.T) {
	ctx := t.Context()
	opts := rez.NewLoggerOptions{
		Level: slog.LevelWarn,
	}
	logger := NewLogger(opts)

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
