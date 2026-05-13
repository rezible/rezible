package http

import (
	"log/slog"
	"net/http"
	"time"
)

type requestLogger struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *requestLogger) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *requestLogger) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}

	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func RequestLoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rec := &requestLogger{
				ResponseWriter: w,
				status:         0,
			}

			next.ServeHTTP(rec, r)

			status := rec.status
			if status == 0 {
				status = http.StatusOK
			}

			level := slog.LevelInfo
			if status >= 500 {
				level = slog.LevelError
			} else if status >= 400 {
				level = slog.LevelWarn
			}

			logger.LogAttrs(
				r.Context(),
				level,
				"http_request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", status),
				slog.Duration("duration", time.Since(start)),
				slog.Int("bytes", rec.bytes),
			)
		})
	}
}
