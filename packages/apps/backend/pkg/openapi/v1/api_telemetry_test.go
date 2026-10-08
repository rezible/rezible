package v1

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestRejectedRequestSpanKeepsOperation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		headers []any
		detail  string
	}{
		{"identity rejected", nil, "Sign in to continue."},
		{"session rejected", []any{"Cookie: " + AppAuthSessionCookieName + "=expired"}, "Sign in to continue."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spans := tracetest.NewSpanRecorder()
			ctx, span := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans)).
				Tracer("test").Start(t.Context(), http.MethodGet)

			response := newSecuredAPI(t, cookieSecurity{}).GetCtx(ctx, securedPath, tc.headers...)
			span.End()

			require.Equal(t, http.StatusUnauthorized, response.Code)
			require.Equal(t, tc.detail, decodeErrorResponse(t, response).Detail)
			ended := spans.Ended()
			require.Len(t, ended, 1)
			require.Equal(t, securedOperation.OperationID, ended[0].Name())
			require.Contains(t, ended[0].Attributes(), attribute.String("http.route", securedOperation.Path))
		})
	}
}
