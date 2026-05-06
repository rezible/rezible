package telemetry

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/lmittmann/tint"
)

type (
	Logger     = slog.Logger
	LogLevel   = slog.Level
	LogLeveler = slog.Leveler
	LogAttr    = slog.Attr
)

func makeSlogHandler(w io.Writer, cfg Config) slog.Handler {
	opts := &slog.HandlerOptions{
		AddSource:   cfg.Logging.AddSource,
		Level:       cfg.getLogLevel(),
		ReplaceAttr: nil,
	}
	if cfg.Logging.Json {
		return slog.NewJSONHandler(w, opts)
	}
	if !cfg.Logging.Color {
		return slog.NewTextHandler(w, opts)
	}
	return tint.NewHandler(w, &tint.Options{
		Level:      opts.Level,
		TimeFormat: time.Kitchen,
	})
}

type loggerContextKey struct{}

type loggerOptions struct {
	parent *slog.Logger
	level  LogLeveler
	attrs  []LogAttr
	args   []any
	groups []string
}

type LoggerOption func(*loggerOptions)

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		logger = slog.Default()
	}
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

func ContextWithLoggerOptions(ctx context.Context, opts ...LoggerOption) context.Context {
	return ContextWithLogger(ctx, NewLogger(ctx, opts...))
}

func LoggerFromContext(ctx context.Context, opts ...LoggerOption) *Logger {
	return NewLogger(ctx, opts...)
}

func NewLogger(ctx context.Context, opts ...LoggerOption) *Logger {
	cfg := loggerOptions{}
	for _, opt := range opts {
		opt(&cfg)
	}

	logger := cfg.parent
	if logger == nil && ctx != nil {
		logger, _ = ctx.Value(loggerContextKey{}).(*Logger)
	}
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.level != nil {
		logger = slog.New(levelHandler{
			base:  logger.Handler(),
			level: cfg.level,
		})
	}
	if len(cfg.attrs) > 0 {
		logger = logger.With(slogAttrsToArgs(cfg.attrs)...)
	}
	if len(cfg.args) > 0 {
		logger = logger.With(cfg.args...)
	}
	for _, group := range cfg.groups {
		logger = logger.WithGroup(group)
	}
	return logger
}

func WithParentLogger(logger *slog.Logger) LoggerOption {
	return func(opts *loggerOptions) {
		opts.parent = logger
	}
}

func WithMinLogLevel(level LogLeveler) LoggerOption {
	return func(opts *loggerOptions) {
		opts.level = level
	}
}

func WithLogAttrs(attrs ...LogAttr) LoggerOption {
	return func(opts *loggerOptions) {
		opts.attrs = append(opts.attrs, attrs...)
	}
}

func WithLogValues(args ...any) LoggerOption {
	return func(opts *loggerOptions) {
		opts.args = append(opts.args, args...)
	}
}

func WithLogGroup(name string) LoggerOption {
	return func(opts *loggerOptions) {
		if name != "" {
			opts.groups = append(opts.groups, name)
		}
	}
}

func WithLogPackage(name string) LoggerOption {
	if name == "" {
		name = "unknown"
	}
	return WithLogValues("package", name)
}

type levelHandler struct {
	base  slog.Handler
	level LogLeveler
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

func slogAttrsToArgs(attrs []LogAttr) []any {
	args := make([]any, 0, len(attrs))
	for _, attr := range attrs {
		args = append(args, attr)
	}
	return args
}
