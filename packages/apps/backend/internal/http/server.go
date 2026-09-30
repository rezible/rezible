package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/execution"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
)

type (
	Server struct {
		logger *slog.Logger

		server *http.Server

		listenerMu    sync.Mutex
		listenerReady atomic.Bool
	}

	UserAuthProvider interface {
		MakeAuthHandler() http.Handler
	}

	WebhookHandlers map[string]http.Handler

	HealthCheckFunc = func(context.Context) map[string]error
)

const (
	healthCheckPath    = "/health"
	readinessCheckPath = "/ready"
)

func NewServer(
	cfg rez.Config,
	userAuth UserAuthProvider,
	v1Api oapiv1.API,
	webhooks WebhookHandlers,
	healthFn HealthCheckFunc,
) (*Server, error) {
	s := &Server{
		logger: slog.Default().WithGroup("http"),
	}

	router := chi.NewMux()

	router.Get(healthCheckPath, s.makeHealthCheckHandler(healthFn))
	router.Get(readinessCheckPath, s.makeReadyCheckHandler())

	webhooksHandler := chi.NewMux()
	for prefix, wh := range webhooks {
		route := ensureSlashPrefix(prefix)
		slog.Debug("mounting webhook handler", "route", route)
		webhooksHandler.Mount(route, wh)
	}
	router.Mount("/webhooks", webhooksHandler)

	router.Mount("/auth", userAuth.MakeAuthHandler())

	router.Mount(oapiv1.VersionPrefix, v1Api.Adapter())

	s.server = s.makeServer(cfg, router)

	return s, nil
}

func (s *Server) makeServer(cfg rez.Config, r *chi.Mux) *http.Server {
	handler := chi.NewRouter()
	handler.Use(s.makeSetRootExecutionContextMiddleware())
	handler.Use(s.makeRequestLoggerMiddleware(cfg.App.DebugMode))
	httpCfg := cfg.HttpServer
	handler.Mount(ensureSlashPrefix(httpCfg.BasePath), http.StripPrefix(httpCfg.BasePath, r))
	return &http.Server{
		Addr:    net.JoinHostPort(httpCfg.Host, httpCfg.Port),
		Handler: handler,
	}
}

func ensureSlashPrefix(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}
	return s
}

func (s *Server) makeSetRootExecutionContextMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			execCtx := execution.NewRootContext(r.Context(), execution.KindAnonymous, execution.SourceHTTP)
			next.ServeHTTP(w, r.WithContext(execCtx))
		})
	}
}

func (s *Server) makeHealthCheckHandler(hc HealthCheckFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checkErrors := hc(r.Context())
		if len(checkErrors) > 0 {
			slog.Debug("health check errors", "errors", checkErrors)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func (s *Server) makeReadyCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.listenerReady.Load() {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) makeRequestLoggerMiddleware(concise bool) func(http.Handler) http.Handler {
	logFormat := httplog.SchemaECS.Concise(concise)
	isDebugHeaderSet := func(r *http.Request) bool {
		return r.Header.Get("Debug") == "reveal-body-logs"
	}

	skipPaths := mapset.NewThreadUnsafeSet(healthCheckPath, readinessCheckPath)
	return httplog.RequestLogger(s.logger, &httplog.Options{
		Level:         slog.LevelInfo,
		Schema:        logFormat,
		RecoverPanics: true,

		Skip: func(req *http.Request, respStatus int) bool {
			if skipPaths.Contains(req.URL.Path) {
				return true
			}
			return respStatus == 404 || respStatus == 405
		},

		LogRequestHeaders:  []string{"Origin"},
		LogResponseHeaders: []string{},

		LogRequestBody:  isDebugHeaderSet,
		LogResponseBody: isDebugHeaderSet,
	})
}

func (s *Server) Run(ctx context.Context, ready chan<- struct{}) error {
	s.server.BaseContext = func(net.Listener) context.Context {
		return ctx
	}

	listener, listenerErr := s.makeListener()
	if listenerErr != nil {
		return fmt.Errorf("listener: %w", listenerErr)
	}
	close(ready)

	slog.Info("HTTP server listening", "addr", s.server.Addr)

	if serveErr := s.server.Serve(listener); !errors.Is(serveErr, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server: %w", serveErr)
	}
	return nil
}

func (s *Server) Handler() http.Handler {
	return s.server.Handler
}

func (s *Server) makeListener() (net.Listener, error) {
	s.listenerMu.Lock()
	defer s.listenerMu.Unlock()
	if s.listenerReady.Load() {
		return nil, fmt.Errorf("HTTP server is already started")
	}
	l, listenerErr := net.Listen("tcp", s.server.Addr)
	if listenerErr != nil {
		return nil, fmt.Errorf("listener: %w", listenerErr)
	}
	s.listenerReady.Store(true)
	return l, nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.listenerMu.Lock()
	defer s.listenerMu.Unlock()
	s.listenerReady.Store(false)
	if s.server != nil {
		slog.Info("HTTP server shutting down")
		if shutdownErr := s.server.Shutdown(ctx); shutdownErr != nil {
			return errors.Join(fmt.Errorf("shutdown HTTP server: %w", shutdownErr), s.server.Close())
		}
	}
	return nil
}

var (
	docsBodyScalar = []byte(`<!doctype html>
<html lang="en">
	<head>
		<title>API Reference</title>
		<meta charset="utf-8" />
		<meta
		name="viewport"
		content="width=device-width, initial-scale=1" />
	</head>
	<body>
		<script id="api-reference" data-url="/api/v1/openapi.json"></script>
		<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
	</body>
</html>`)

	docsBodyStoplight = []byte(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="referrer" content="same-origin" />
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no" />
    <title>API Dev Docs</title>
    <link href="https://unpkg.com/@stoplight/elements/styles.min.css" rel="stylesheet" />
    <script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
  </head>
  <body style="height: 100vh;">
    <elements-api
      apiDescriptionUrl="/api/v1/openapi.json"
      router="hash"
      layout="sidebar"
      tryItCredentialsPolicy="same-origin"
    />
  </body>
</html>`)
)

func serveApiDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if _, wErr := w.Write(docsBodyScalar); wErr != nil {
		slog.Error("failed to write embedded docs body", "error", wErr)
	}
}
