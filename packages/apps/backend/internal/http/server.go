package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/koding/websocketproxy"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/openapi"
	slogchi "github.com/samber/slog-chi"

	"github.com/go-chi/chi/v5"
	"github.com/rezible/rezible/execution"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

type Server struct {
	cfg        Config
	router     *chi.Mux
	logger     *slog.Logger
	httpServer *http.Server
}

func ensureSlashPrefix(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}
	return s
}

type UserAuthSessionService interface {
	Handler() http.Handler
	ExecutionContextMiddleware() func(http.Handler) http.Handler
}

func NewServer(ts rez.TelemetryService, auth UserAuthSessionService, oapiV1Handler oapiv1.Handler, webhookHandlers map[string]http.Handler) (*Server, error) {
	cfg, cfgErr := loadConfig()
	if cfgErr != nil {
		return nil, fmt.Errorf("config error: %w", cfgErr)
	}

	s := Server{
		cfg:    cfg,
		router: chi.NewRouter(),
		logger: slog.Default().WithGroup("http"),
	}

	s.router.Use(s.makeExecutionContextMiddleware())
	s.router.Use(s.makeRequestLoggerMiddleware())

	api := oapiv1.MakeApi(oapiV1Handler, oapiv1.MakeSecurityMiddleware(), oapiv1.MakeAPITelemetryMiddleware(ts))

	if handlerErr := s.mountRequestHandler(auth, api, webhookHandlers); handlerErr != nil {
		return nil, fmt.Errorf("http request handler: %w", handlerErr)
	}

	return &s, nil
}

func (s *Server) makeExecutionContextMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			execCtx := execution.NewRootContext(r.Context(), execution.KindAnonymous, execution.SourceHTTP)
			next.ServeHTTP(w, r.WithContext(execCtx))
		})
	}
}

func (s *Server) makeRequestLoggerMiddleware() func(http.Handler) http.Handler {
	return slogchi.NewWithConfig(s.logger, slogchi.Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelInfo,
		ServerErrorLevel: slog.LevelError,
		WithRequestID:    true,
		WithSpanID:       true,
		WithTraceID:      true,
	})
}

func (s *Server) mountRequestHandler(auth UserAuthSessionService, api openapi.API, webhookHandlers map[string]http.Handler) error {
	r := chi.NewRouter()

	r.Get("/health", s.makeHealthCheckHandler())

	webhooks := chi.NewMux()
	for route, wh := range webhookHandlers {
		webhooks.Mount(ensureSlashPrefix(route), wh)
	}
	r.Mount("/webhooks", webhooks)

	r.Mount("/auth", auth.Handler())

	// routes requiring auth context
	r.Group(func(ar chi.Router) {
		ar.Use(auth.ExecutionContextMiddleware())

		if s.cfg.DocumentsProxy.Enabled {
			ar.Handle("/documents", s.makeDocumentsProxyHandler())
		}
		ar.Mount(oapiv1.VersionPrefix, api.Adapter())
	})

	basePath := s.cfg.BasePath
	s.router.Mount(ensureSlashPrefix(basePath), http.StripPrefix(basePath, r))

	return nil
}

func (s *Server) makeDocumentsProxyHandler() http.Handler {
	headerKey := "X-Rez-Tenant-ID"
	setAuthHeaders := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			exec := execution.GetContext(r.Context())
			tenantId, tenantOk := exec.TenantID()
			if exec.IsAnonymous() || exec.Auth.TokenID != nil || !tenantOk {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			r.Header.Set(headerKey, fmt.Sprintf("%d", tenantId))
			next.ServeHTTP(w, r)
		})
	}
	proxy := websocketproxy.NewProxy(s.cfg.DocumentsProxy.serverUrl)
	proxy.Director = func(r *http.Request, h http.Header) {
		h.Set(headerKey, r.Header.Get(headerKey))
	}
	return chi.Chain(setAuthHeaders).Handler(proxy)
}

func (s *Server) makeHealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) Start(baseCtx context.Context) error {
	s.httpServer = &http.Server{
		Addr:    net.JoinHostPort(s.cfg.Host, s.cfg.Port),
		Handler: s.router,
		BaseContext: func(l net.Listener) context.Context {
			return baseCtx
		},
	}

	slog.Info("HTTP server listening", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server error: %w", err)
	}
	slog.Info("Stopped HTTP server")
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
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
