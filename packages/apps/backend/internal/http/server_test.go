package http

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
)

func TestRequestLogsOmitWebhookTokens(t *testing.T) {
	const token = "c2VjcmV0LXdlYmhvb2stdG9rZW4tdmFsdWUtZm9yLXRlc3Rpbmc"
	logs := &bytes.Buffer{}
	s := &Server{logger: slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug}))}

	router := chi.NewMux()
	router.Post("/webhooks/alertmanager/{token}", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid payload", http.StatusBadRequest)
	})
	router.Post("/v1/integrations/installations/{id}/webhook-token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"url":"https://app.test/api/webhooks/alertmanager/` + token + `"}}`))
	})
	cfg := rez.Config{HttpServer: rez.HttpServerConfig{BasePath: "/api"}}
	handler := s.makeServer(cfg, router).Handler

	serve := func(path string) int {
		request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"version":"3"}`))
		request.Header.Set("Debug", "reveal-body-logs")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder.Code
	}

	require.Equal(t, http.StatusBadRequest, serve("/api/webhooks/alertmanager/"+token))
	require.Equal(t, http.StatusOK, serve("/api/v1/integrations/installations/b1a5/webhook-token"))

	require.Contains(t, logs.String(), "webhook-token", "the issuance request is logged")
	require.NotContains(t, logs.String(), token)
}
