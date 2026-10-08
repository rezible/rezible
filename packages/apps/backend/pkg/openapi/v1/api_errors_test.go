package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
)

type testErrorInput struct {
	Body struct {
		Name string `json:"name" minLength:"3"`
	}
}

var testErrorOperation = huma.Operation{
	OperationID: "fail",
	Method:      http.MethodPost,
	Path:        "/fail",
}

// newFailingAPI registers one operation that returns operationErr through Error.
func newFailingAPI(t *testing.T, operationErr error) humatest.TestAPI {
	_, api := humatest.New(t)
	huma.Register(api, testErrorOperation, func(ctx context.Context, _ *testErrorInput) (*EmptyResponse, error) {
		return nil, Error(ctx, "fail operation", operationErr)
	})
	return api
}

func decodeErrorResponse(t *testing.T, response *httptest.ResponseRecorder) errorModel {
	t.Helper()
	var body errorModel
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body), response.Body.String())
	require.Equal(t, response.Code, body.Status)
	return body
}

type capturedLogs struct {
	mu      sync.Mutex
	records []slog.Record
}

func captureLogs(t *testing.T) *capturedLogs {
	previous := slog.Default()
	logs := &capturedLogs{}
	slog.SetDefault(slog.New(logs))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return logs
}

func (c *capturedLogs) Enabled(context.Context, slog.Level) bool { return true }

func (c *capturedLogs) Handle(_ context.Context, record slog.Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records = append(c.records, record)
	return nil
}

func (c *capturedLogs) WithAttrs([]slog.Attr) slog.Handler { return c }
func (c *capturedLogs) WithGroup(string) slog.Handler      { return c }

func (c *capturedLogs) requireOne(t *testing.T, level slog.Level, code errs.Code) {
	t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()
	require.Len(t, c.records, 1)
	require.Equal(t, level, c.records[0].Level)
	var loggedCode any
	c.records[0].Attrs(func(attr slog.Attr) bool {
		if attr.Key == "code" {
			loggedCode = attr.Value.Any()
		}
		return true
	})
	require.Equal(t, code, loggedCode)
}

func TestInternalErrorsDoNotReachResponses(t *testing.T) {
	for _, tc := range []struct {
		name   string
		opErr  error
		hidden []string
	}{
		{
			name:   "plain",
			opErr:  errors.New("dial tcp 10.0.0.3:5432: connection refused"),
			hidden: []string{"10.0.0.3", "connection refused"},
		},
		{
			name:   "wrapped sql",
			opErr:  fmt.Errorf("list situations for tenant 42: %w", errors.New(`pq: relation "situations" does not exist`)),
			hidden: []string{"tenant 42", "relation", "situations"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logs := captureLogs(t)
			response := newFailingAPI(t, tc.opErr).Post("/fail", map[string]any{"name": "valid"})

			require.Equal(t, http.StatusInternalServerError, response.Code)
			for _, text := range tc.hidden {
				require.NotContains(t, response.Body.String(), text)
			}
			body := decodeErrorResponse(t, response)
			require.Equal(t, errs.CodeInternal, body.Code)
			require.Empty(t, body.Errors)
			logs.requireOne(t, slog.LevelError, errs.CodeInternal)
		})
	}
}

func TestSentinelResponses(t *testing.T) {
	for _, tc := range []struct {
		sentinel error
		status   int
		code     errs.Code
		detail   string
	}{
		{errs.ErrInvalidInput, http.StatusBadRequest, errs.CodeInvalidInput, ""},
		{errs.ErrUnprocessableInput, http.StatusUnprocessableEntity, errs.CodeUnprocessable, ""},
		{errs.ErrAuthSessionMissing, http.StatusUnauthorized, errs.CodeUnauthenticated, ""},
		{errs.ErrAuthSessionExpired, http.StatusUnauthorized, errs.CodeUnauthenticated, ""},
		{errs.ErrAuthSessionInvalid, http.StatusUnauthorized, errs.CodeUnauthenticated, ""},
		{errs.ErrForbidden, http.StatusForbidden, errs.CodeForbidden, ""},
		{errs.ErrNotFound, http.StatusNotFound, errs.CodeNotFound, ""},
		{errs.ErrConflict, http.StatusConflict, errs.CodeConflict, ""},
		{errs.ErrRateLimited, http.StatusTooManyRequests, errs.CodeRateLimited, ""},
		{errs.ErrNotImplemented, http.StatusNotImplemented, errs.CodeNotImplemented, ""},
		{errs.ErrTenantContextMissing, http.StatusInternalServerError, errs.CodeInternal, ""},
	} {
		t.Run(string(tc.code)+"/"+tc.sentinel.Error(), func(t *testing.T) {
			logs := captureLogs(t)
			opErr := fmt.Errorf("update incident 7: %w", tc.sentinel)
			response := newFailingAPI(t, opErr).Post("/fail", map[string]any{"name": "valid"})

			require.Equal(t, tc.status, response.Code)
			require.NotContains(t, response.Body.String(), "incident 7")
			body := decodeErrorResponse(t, response)
			require.Equal(t, tc.code, body.Code)
			if tc.detail == "" {
				tc.detail = errs.PublicMessage(errs.New(tc.code, ""))
			}
			require.Equal(t, tc.detail, body.Detail)

			level := slog.LevelDebug
			if tc.status >= http.StatusInternalServerError {
				level = slog.LevelError
			}
			logs.requireOne(t, level, tc.code)
		})
	}
}

func TestEntErrorResponses(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		response := newFailingAPI(t, fmt.Errorf("get team: %w", &ent.NotFoundError{})).Post("/fail", map[string]any{"name": "valid"})
		require.Equal(t, http.StatusNotFound, response.Code)
		require.Equal(t, errs.CodeNotFound, decodeErrorResponse(t, response).Code)
	})

	t.Run("constraint with field detail", func(t *testing.T) {
		opErr := errors.Join(errors.New(`create team: duplicate key violates unique constraint "teams_name_key"`), &ent.ConstraintError{})
		response := newFailingAPI(t, opErr).Post("/fail", map[string]any{"name": "valid"})

		require.Equal(t, http.StatusBadRequest, response.Code)
		require.NotContains(t, response.Body.String(), "teams_name_key")
		body := decodeErrorResponse(t, response)
		require.Equal(t, errs.CodeInvalidInput, body.Code)
		require.Len(t, body.Errors, 1)
		require.Equal(t, "name", body.Errors[0].Location)
		require.Equal(t, "Name already exists", body.Errors[0].Message)
	})

	t.Run("constraint on another field", func(t *testing.T) {
		opErr := errors.Join(errors.New(`create team: duplicate key violates unique constraint "teams_slug_key"`), &ent.ConstraintError{})
		response := newFailingAPI(t, opErr).Post("/fail", map[string]any{"name": "valid"})

		require.Equal(t, http.StatusConflict, response.Code)
		require.Equal(t, errs.CodeConflict, decodeErrorResponse(t, response).Code)
	})
}

func TestHumaValidationErrorKeepsDetails(t *testing.T) {
	logs := captureLogs(t)
	response := newFailingAPI(t, nil).Post("/fail", map[string]any{"name": "a"})

	require.Equal(t, http.StatusUnprocessableEntity, response.Code)
	body := decodeErrorResponse(t, response)
	require.Equal(t, errs.CodeInvalidInput, body.Code)
	require.NotEmpty(t, body.Errors)
	require.Equal(t, "body.name", body.Errors[0].Location)
	logs.requireOne(t, slog.LevelDebug, errs.CodeInvalidInput)
}

// foreignStatusError is a status error not built by this package.
type foreignStatusError struct {
	status int
	text   string
}

func (e foreignStatusError) Error() string  { return e.text }
func (e foreignStatusError) GetStatus() int { return e.status }

// failingSecurity rejects every request with its error.
type failingSecurity struct{ rejectErr error }

func (s failingSecurity) CreateRequestSecurityContext(http.ResponseWriter, *http.Request) (context.Context, error) {
	return nil, s.rejectErr
}

func (s failingSecurity) VerifyRequestSecurity(context.Context, OperationSecurityOptions) error {
	return nil
}

// securedHandler implements no operations; the tests' security providers reject each request before one runs.
type securedHandler struct {
	Handler
	security SecurityProvider
}

func (h securedHandler) CreateRequestSecurityContext(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	return h.security.CreateRequestSecurityContext(w, r)
}

func (h securedHandler) VerifyRequestSecurity(ctx context.Context, opts OperationSecurityOptions) error {
	return h.security.VerifyRequestSecurity(ctx, opts)
}

// securedOperation is an operation that requires a session; securedPath is its request path.
var (
	securedOperation = ListSituations
	securedPath      = VersionPrefix + securedOperation.Path
)

func newSecuredAPI(t *testing.T, provider SecurityProvider) humatest.TestAPI {
	return humatest.Wrap(t, MakeApi(securedHandler{security: provider}))
}

func TestPrebuiltStatusErrorsLeakNothing(t *testing.T) {
	foreignModel := &huma.ErrorModel{Status: http.StatusBadRequest, Detail: "pq: duplicate key in tenant 7"}
	foreignModel.Add(errors.New("pq: duplicate key in tenant 7"))

	for _, tc := range []struct {
		name   string
		opErr  error
		status int
		code   errs.Code
	}{
		{"huma model", foreignModel, http.StatusBadRequest, errs.CodeInvalidInput},
		{"other type", fmt.Errorf("save: %w", foreignStatusError{status: http.StatusConflict, text: "pq: duplicate key in tenant 7"}), http.StatusConflict, errs.CodeConflict},
		{"huma 5xx constructor", huma.Error500InternalServerError("pq: duplicate key in tenant 7"), http.StatusInternalServerError, errs.CodeInternal},
	} {
		boundaries := map[string]func(t *testing.T) *httptest.ResponseRecorder{
			"handler": func(t *testing.T) *httptest.ResponseRecorder {
				return newFailingAPI(t, tc.opErr).Post("/fail", map[string]any{"name": "valid"})
			},
			"security middleware": func(t *testing.T) *httptest.ResponseRecorder {
				return newSecuredAPI(t, failingSecurity{rejectErr: tc.opErr}).Get(securedPath)
			},
		}
		for boundary, request := range boundaries {
			t.Run(tc.name+"/"+boundary, func(t *testing.T) {
				logs := captureLogs(t)
				response := request(t)

				require.Equal(t, tc.status, response.Code)
				require.NotContains(t, response.Body.String(), "tenant 7")
				body := decodeErrorResponse(t, response)
				require.Equal(t, tc.code, body.Code)
				require.Equal(t, errs.PublicMessage(errs.New(tc.code, "")), body.Detail)
				require.Empty(t, body.Errors)
				level := slog.LevelDebug
				if tc.status >= http.StatusInternalServerError {
					level = slog.LevelError
				}
				logs.requireOne(t, level, tc.code)
			})
		}
	}
}

func TestWrittenConstructorErrorIsLoggedOnce(t *testing.T) {
	logs := captureLogs(t)
	_, api := humatest.New(t)
	api.UseMiddleware(func(c huma.Context, _ func(huma.Context)) {
		_ = huma.WriteErr(api, c, http.StatusConflict, "", huma.Error409Conflict("A team with that name exists."))
	})
	huma.Register(api, huma.Operation{OperationID: "get-team", Method: http.MethodGet, Path: "/team"},
		func(context.Context, *struct{}) (*EmptyResponse, error) { return &EmptyResponse{}, nil })

	response := api.Get("/team")

	require.Equal(t, http.StatusConflict, response.Code)
	require.Equal(t, "A team with that name exists.", decodeErrorResponse(t, response).Detail)
	logs.requireOne(t, slog.LevelDebug, errs.CodeConflict)
}

func TestErrorAddsAttributesToSpan(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tracer := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)).Tracer("test")
	ctx, span := tracer.Start(t.Context(), "request")

	cause := errs.Wrap(errors.New("situation closed"), errs.CodeConflict, "", attribute.String("situation_id", "s-1"))
	_ = Error(ctx, "raise situation", fmt.Errorf("raise: %w", cause))
	span.End()

	require.Equal(t, []attribute.KeyValue{
		attribute.String("code", "conflict"),
		attribute.String("situation_id", "s-1"),
	}, recorder.Ended()[0].Attributes())
}

// cookieSecurity treats any session cookie as expired, and a request without one as anonymous.
type cookieSecurity struct{}

func (cookieSecurity) CreateRequestSecurityContext(_ http.ResponseWriter, r *http.Request) (context.Context, error) {
	if _, cookieErr := r.Cookie(AppAuthSessionCookieName); cookieErr == nil {
		return nil, errs.ErrAuthSessionExpired
	}
	return execution.NewRootContext(r.Context(), execution.KindAnonymous, execution.SourceHTTP), nil
}

func (cookieSecurity) VerifyRequestSecurity(ctx context.Context, _ OperationSecurityOptions) error {
	if execution.GetContext(ctx).IsAnonymous() {
		return errs.ErrAuthSessionMissing
	}
	return nil
}

func TestSecurityMiddlewareErrorResponses(t *testing.T) {
	testAPI := newSecuredAPI(t, cookieSecurity{})

	for _, tc := range []struct {
		name   string
		args   []any
		detail string
	}{
		{"no session cookie", nil, "Sign in to continue."},
		{"expired session", []any{"Cookie: " + AppAuthSessionCookieName + "=expired"}, "Sign in to continue."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logs := captureLogs(t)
			response := testAPI.Get(securedPath, tc.args...)

			require.Equal(t, http.StatusUnauthorized, response.Code)
			body := decodeErrorResponse(t, response)
			require.Equal(t, tc.detail, body.Detail)
			require.Equal(t, errs.CodeUnauthenticated, body.Code)
			logs.requireOne(t, slog.LevelDebug, errs.CodeUnauthenticated)
		})
	}
}

func TestEveryCodeHasAStatus(t *testing.T) {
	for _, code := range errs.Codes() {
		require.Contains(t, codeStatuses, code, "a code without a status is answered as a 500")
	}
}
