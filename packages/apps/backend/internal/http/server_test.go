package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	rez "github.com/rezible/rezible"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/test"
)

type ServerSuite struct {
	test.Suite
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, &ServerSuite{Suite: test.NewSuite()})
}

func (s *ServerSuite) newServer() (*Server, *bytes.Buffer) {
	logs := &bytes.Buffer{}
	server := &Server{
		logger: slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
	}
	return server, logs
}

func (s *ServerSuite) TestRequestLogsOmitWebhookTokens() {
	const token = "c2VjcmV0LXdlYmhvb2stdG9rZW4tdmFsdWUtZm9yLXRlc3Rpbmc"
	server, logs := s.newServer()

	rejectDelivery := chi.NewRouter()
	rejectDelivery.Post("/{token}", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid payload", http.StatusBadRequest)
	})
	webhooks := WebhookHandlers{
		"alertmanager": rejectDelivery,
		"webhook":      rejectDelivery,
		"github":       rejectDelivery,
	}

	router := chi.NewMux()
	router.Mount(webhooksPath, server.makeWebhooksRouter(webhooks))
	router.Post("/v1/integrations/installations/{id}/webhook-token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"url":"https://app.test/api/webhooks/alertmanager/` + token + `"}}`))
	})
	cfg := rez.Config{
		HttpServer: rez.HttpServerConfig{BasePath: "/api"},
	}
	handler := server.makeServer(cfg, router).Handler

	serve := func(path string) int {
		request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"version":"3"}`))
		request.Header.Set("Debug", "reveal-body-logs")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder.Code
	}

	for name := range webhooks {
		s.Equal(http.StatusBadRequest, serve("/api/webhooks/"+name+"/"+token), name)
		s.Contains(logs.String(), "path=/webhooks/"+name+" status=400", "the %s request is logged by its handler's path", name)
	}
	s.Equal(http.StatusOK, serve("/api/v1/integrations/installations/b1a5/webhook-token"))

	s.Contains(logs.String(), "webhook-token", "the issuance request is logged")
	s.NotContains(logs.String(), token)
}

func (s *ServerSuite) TestPanickingWebhookRequestIsLoggedWithoutItsToken() {
	const token = "c2VjcmV0LXdlYmhvb2stdG9rZW4tdmFsdWUtZm9yLXRlc3Rpbmc"
	server, logs := s.newServer()

	panicking := chi.NewRouter()
	panicking.Post("/{token}", func(w http.ResponseWriter, r *http.Request) {
		panic("delivery failed for " + r.URL.Path)
	})
	webhooks := WebhookHandlers{
		"webhook": panicking,
	}
	router := chi.NewMux()
	router.Mount(webhooksPath, server.makeWebhooksRouter(webhooks))
	cfg := rez.Config{
		HttpServer: rez.HttpServerConfig{BasePath: "/api"},
	}
	handler := server.makeServer(cfg, router).Handler

	request := httptest.NewRequest(http.MethodPost, "/api/webhooks/webhook/"+token, strings.NewReader(`{}`))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	s.Equal(http.StatusInternalServerError, recorder.Code)
	s.Contains(logs.String(), "path=/webhooks/webhook panic=string")
	s.Contains(logs.String(), "status=500")
	s.NotContains(logs.String(), token)
}

func (s *ServerSuite) TestHealthCheckReportsOnlyFailingServices() {
	server, logs := s.newServer()
	results := map[string]error{
		"*db.AlertService":     nil,
		"*db.SituationService": nil,
	}
	handler := server.makeHealthCheckHandler(func(context.Context) map[string]error {
		return results
	})
	serve := func() int {
		recorder := httptest.NewRecorder()
		handler(recorder, httptest.NewRequest(http.MethodGet, healthCheckPath, nil))
		return recorder.Code
	}

	s.Equal(http.StatusOK, serve(), "services without errors are healthy")
	s.Empty(logs.String())

	results["*postgres.ConnectionPool"] = errors.New("connection refused")
	s.Equal(http.StatusInternalServerError, serve())
	s.Contains(logs.String(), "postgres.ConnectionPool")
	s.Contains(logs.String(), "connection refused")
	s.NotContains(logs.String(), "AlertService", "healthy services are not logged")
}

func (s *ServerSuite) TestRequestSpanReachesResponseAndLogs() {
	spans := tracetest.NewSpanRecorder()
	previousProvider := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans)))
	s.T().Cleanup(func() { otel.SetTracerProvider(previousProvider) })
	// The server takes its logger from the default when it is built.
	logs := &bytes.Buffer{}
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(traceIDHandler{Handler: slog.NewJSONHandler(logs, nil)}))
	s.T().Cleanup(func() { slog.SetDefault(previousLogger) })

	api := humago.NewWithPrefix(http.NewServeMux(), oapiv1.VersionPrefix, huma.DefaultConfig("Test", "1"))
	huma.Register(api, huma.Operation{OperationID: "get-thing", Method: http.MethodGet, Path: "/things/{id}"},
		func(ctx context.Context, _ *struct {
			ID string `path:"id"`
		}) (*struct{}, error) {
			return nil, nil
		})
	rejectDelivery := chi.NewRouter()
	rejectDelivery.Post("/{token}", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid payload", http.StatusBadRequest)
	})
	cfg := rez.Config{
		HttpServer: rez.HttpServerConfig{BasePath: "/api"},
	}
	webhooks := WebhookHandlers{"webhook": rejectDelivery}
	server, serverErr := NewServer(cfg, noUserAuth{}, api, webhooks, nil)
	s.Require().NoError(serverErr)
	handler := server.Handler()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/things/b1a5", nil))
	s.Equal(http.StatusNoContent, recorder.Code)

	ended := spans.Ended()
	s.Require().Len(ended, 1)
	traceID := ended[0].SpanContext().TraceID().String()
	s.Contains(recorder.Header().Get("traceparent"), traceID)

	var requestLog map[string]any
	s.Require().NoError(json.Unmarshal(logs.Bytes(), &requestLog), logs.String())
	s.Equal("http", requestLog["component"])
	s.Equal(traceID, requestLog["trace_id"], "the request log line carries the trace ID at the top level")

	// A webhook URL carries a secret token in its path, so webhook requests are not traced.
	webhookRecorder := httptest.NewRecorder()
	handler.ServeHTTP(webhookRecorder, httptest.NewRequest(http.MethodPost, "/api/webhooks/webhook/secret-token", nil))
	s.Equal(http.StatusBadRequest, webhookRecorder.Code)
	s.Len(spans.Ended(), 1)
	s.Empty(webhookRecorder.Header().Get("traceparent"))
}

type noUserAuth struct{}

func (noUserAuth) MakeAuthHandler() http.Handler { return http.NotFoundHandler() }

// traceIDHandler adds the trace ID of a record's context, as the production log handler in
// internal/opentelemetry does. Attributes added while handling land in any group the logger opened, so a
// grouped logger would nest the trace ID.
type traceIDHandler struct {
	slog.Handler
}

func (h traceIDHandler) Handle(ctx context.Context, record slog.Record) error {
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		record.AddAttrs(slog.String("trace_id", spanCtx.TraceID().String()))
	}
	return h.Handler.Handle(ctx, record)
}

func (h traceIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return traceIDHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h traceIDHandler) WithGroup(name string) slog.Handler {
	return traceIDHandler{Handler: h.Handler.WithGroup(name)}
}

func (s *ServerSuite) TestServerErrorsAreLoggedOnceAtError() {
	handler := s.newFailingAPIServer(slog.LevelDebug)
	for _, path := range []string{"/api/v1/fail", "/api/v1/explode"} {
		logs := s.serve(handler, path)
		s.Equal(1, logs.count(slog.LevelError), "%s is logged once at Error", path)
		s.Equal(1, logs.count(slog.LevelInfo), "%s has one access log line at Info", path)
	}
}

func (s *ServerSuite) TestAccessLogKeepsConfiguredLevel() {
	handler := s.newFailingAPIServer(slog.LevelWarn)
	logs := s.serve(handler, "/api/v1/fail")
	s.Equal([]slog.Level{slog.LevelError}, logs.levels, "only the error is logged; its access line is below Warn")
}

// serve sends a GET to path, which must fail with 500, and returns what was logged.
func (s *ServerSuite) serve(handler *failingAPIServer, path string) *levelRecorder {
	handler.logs.reset()
	recorder := httptest.NewRecorder()
	handler.server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	s.Equal(http.StatusInternalServerError, recorder.Code, path)
	return handler.logs
}

type failingAPIServer struct {
	server *Server
	logs   *levelRecorder
}

// newFailingAPIServer serves an API whose operations fail with an error and a panic, logging at minLevel.
func (s *ServerSuite) newFailingAPIServer(minLevel slog.Level) *failingAPIServer {
	// The server takes its logger from the default when it is built.
	logs := &levelRecorder{minLevel: minLevel}
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(logs))
	s.T().Cleanup(func() { slog.SetDefault(previousLogger) })

	api := humago.NewWithPrefix(http.NewServeMux(), oapiv1.VersionPrefix, huma.DefaultConfig("Test", "1"))
	huma.Register(api, huma.Operation{OperationID: "fail", Method: http.MethodGet, Path: "/fail"},
		func(ctx context.Context, _ *struct{}) (*struct{}, error) {
			return nil, oapiv1.Error(ctx, "fail operation", errors.New("connection refused"))
		})
	huma.Register(api, huma.Operation{OperationID: "explode", Method: http.MethodGet, Path: "/explode"},
		func(context.Context, *struct{}) (*struct{}, error) {
			panic("handler exploded")
		})
	cfg := rez.Config{
		HttpServer: rez.HttpServerConfig{BasePath: "/api"},
	}
	server, serverErr := NewServer(cfg, noUserAuth{}, api, WebhookHandlers{}, nil)
	s.Require().NoError(serverErr)
	return &failingAPIServer{server: server, logs: logs}
}

// levelRecorder records the level of each record at or above minLevel.
type levelRecorder struct {
	minLevel slog.Level
	mu       sync.Mutex
	levels   []slog.Level
}

func (r *levelRecorder) Enabled(_ context.Context, level slog.Level) bool { return level >= r.minLevel }

func (r *levelRecorder) Handle(_ context.Context, record slog.Record) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.levels = append(r.levels, record.Level)
	return nil
}

func (r *levelRecorder) WithAttrs([]slog.Attr) slog.Handler { return r }
func (r *levelRecorder) WithGroup(string) slog.Handler      { return r }

func (r *levelRecorder) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.levels = nil
}

func (r *levelRecorder) count(level slog.Level) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, l := range r.levels {
		if l == level {
			n++
		}
	}
	return n
}
