package telemetry

import (
	"context"
	"log/slog"

	rez "github.com/rezible/rezible"
)

type loggerContextKey struct{}

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		logger = slog.Default()
	}
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	opts := rez.LoggerOptions{}
	if ctx != nil {
		if parent, ok := ctx.Value(loggerContextKey{}).(*slog.Logger); ok {
			opts.Parent = parent
		}
	}
	return NewLogger(opts)
}

func NewLogger(opts rez.LoggerOptions) *slog.Logger {
	logger := opts.Parent
	if logger == nil {
		logger = slog.Default()
	}
	if opts.Level != nil {
		logger = slog.New(levelHandler{
			base:  logger.Handler(),
			level: opts.Level,
		})
	}
	if len(opts.Attrs) > 0 {
		args := make([]any, 0, len(opts.Attrs))
		for _, attr := range opts.Attrs {
			args = append(args, attr)
		}
		logger = logger.With(args...)
	}
	for _, group := range opts.Groups {
		logger = logger.WithGroup(group)
	}
	return logger
}

type levelHandler struct {
	base  slog.Handler
	level slog.Leveler
}

func (h levelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level.Level() && h.base.Enabled(ctx, level)
}

func (h levelHandler) Handle(ctx context.Context, record slog.Record) error {
	return h.base.Handle(ctx, record)
}

func (h levelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return levelHandler{
		base:  h.base.WithAttrs(attrs),
		level: h.level,
	}
}

func (h levelHandler) WithGroup(name string) slog.Handler {
	return levelHandler{
		base:  h.base.WithGroup(name),
		level: h.level,
	}
}
