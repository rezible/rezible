package river

import (
	"context"
	"log/slog"
)

type logger struct {
	base  slog.Handler
	level slog.Level
}

func (l logger) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= l.level && l.base.Enabled(ctx, level)
}

func (l logger) Handle(ctx context.Context, record slog.Record) error {
	return l.base.Handle(ctx, record)
}

func (l logger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return l.base.WithAttrs(attrs)
}

func (l logger) WithGroup(name string) slog.Handler {
	return l.base.WithGroup(name)
}
