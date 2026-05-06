package v1

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/rezible/rezible/telemetry"
)

func makeAPITelemetryMiddleware() func(huma.Context, func(huma.Context)) {
	m := telemetry.DefaultMeter()
	requests := telemetry.Int64CounterInstrument(m, "rezible.backend.http.server.requests", "HTTP requests handled by the backend")
	requestSeconds := telemetry.Float64HistogramInstrument(m, "rezible.backend.http.server.duration", "HTTP request duration", "s")

	return func(ctx huma.Context, next func(huma.Context)) {
		start := time.Now()

		next(ctx)

		status := ctx.Status()
		if status == 0 {
			status = http.StatusOK
		}
		op := ctx.Operation()
		route := "unknown"
		operationID := "unknown"
		if op != nil {
			route = op.Path
			operationID = op.OperationID
		}

		attrs := []telemetry.KeyValue{
			telemetry.StringAttr("http.request.method", ctx.Method()),
			telemetry.StringAttr("http.route", telemetry.NormalizeLabel(route)),
			telemetry.IntAttr("http.response.status_code", status),
			telemetry.StringAttr("rezible.operation_id", telemetry.NormalizeLabel(operationID)),
		}
		requests.Add(ctx.Context(), 1, telemetry.WithAttributes(attrs...))
		requestSeconds.Record(ctx.Context(), time.Since(start).Seconds(), telemetry.WithAttributes(attrs...))
	}
}
