package v1

import (
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	rez "github.com/rezible/rezible"
)

func MakeAPITelemetryMiddleware(ts rez.TelemetryService) func(huma.Context, func(huma.Context)) {
	m := ts.DefaultMeter()
	requests, requestsErr := m.Int64Counter("rezible.backend.http.server.requests", metric.WithDescription("HTTP requests handled by the backend"))
	requestSeconds, requestSecondsErr := m.Float64Histogram("rezible.backend.http.server.duration", metric.WithDescription("HTTP request duration"), metric.WithUnit("s"))
	if telErr := errors.Join(requestsErr, requestSecondsErr); telErr != nil {
		panic("telemetry error: " + telErr.Error())
	}

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

		attrs := []attribute.KeyValue{
			attribute.String("http.request.method", ctx.Method()),
			attribute.String("http.route", route),
			attribute.Int("http.response.status_code", status),
			attribute.String("rezible.operation_id", operationID),
		}
		requests.Add(ctx.Context(), 1, metric.WithAttributes(attrs...))
		requestSeconds.Record(ctx.Context(), time.Since(start).Seconds(), metric.WithAttributes(attrs...))
	}
}
