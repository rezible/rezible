package http

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
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
