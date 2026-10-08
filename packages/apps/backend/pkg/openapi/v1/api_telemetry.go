package v1

import (
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/openapi"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// nameOperationSpan names the request's span after the API operation and sets its route before security
// runs, so a rejected request's span is named too. The security middleware adds the identity.
func nameOperationSpan(ctx huma.Context, next func(huma.Context)) {
	if op := ctx.Operation(); op != nil {
		span := trace.SpanFromContext(ctx.Context())
		span.SetName(op.OperationID)
		span.SetAttributes(attribute.String("http.route", op.Path))
		span.SetAttributes(execution.Attrs(ctx.Context())...)
	}
	next(ctx)
}

// makeMetricsMiddleware records the API metrics.
func makeMetricsMiddleware() openapi.Middleware {
	m := otel.Meter("github.com/rezible/rezible")
	requests, requestsErr := m.Int64Counter("rezible.backend.http.server.requests",
		metric.WithDescription("HTTP requests handled by the backend"))
	requestSeconds, requestSecondsErr := m.Float64Histogram("rezible.backend.http.server.duration",
		metric.WithDescription("HTTP request duration"),
		metric.WithUnit("s"))
	if telErr := errors.Join(requestsErr, requestSecondsErr); telErr != nil {
		panic("telemetry middleware: " + telErr.Error())
	}

	return func(ctx huma.Context, next func(huma.Context)) {
		start := time.Now()

		next(ctx)

		op := ctx.Operation()
		route := "unknown"
		operationID := "unknown"
		if op != nil {
			route = op.Path
			operationID = op.OperationID
		}
		status := ctx.Status()
		if status == 0 {
			status = http.StatusOK
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
