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
	"github.com/rezible/rezible/internal/http/api/v1"
	"github.com/rezible/rezible/internal/http/oidc"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/integrations"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

type Server struct {
	cfg        Config
	router     *chi.Mux
	httpServer *http.Server
}

func ensureSlashPrefix(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}
	return s
}

func NewServer(ctx context.Context, svcs *rez.Services) (*Server, error) {
	cfg, cfgErr := loadConfig()
	if cfgErr != nil {
		return nil, fmt.Errorf("config error: %w", cfgErr)
	}
	s := Server{cfg: cfg}

	s.router = chi.NewRouter()
	s.router.Use(s.makeExecutionContextMiddleware())
	s.router.Use(s.makeLoggerMiddleware())

	handler, handlerErr := s.makeRequestHandler(ctx, svcs)
	if handlerErr != nil {
		return nil, fmt.Errorf("http request handler: %w", handlerErr)
	}
	base := ensureSlashPrefix(cfg.BasePath)
	s.router.Mount(base, http.StripPrefix(cfg.BasePath, handler))

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

func (s *Server) makeLoggerMiddleware() func(http.Handler) http.Handler {
	return httplog.RequestLogger(slog.Default(), &httplog.Options{
		Level:         slog.LevelInfo,
		Schema:        httplog.SchemaOTEL,
		RecoverPanics: true,
		Skip: func(r *http.Request, respStatus int) bool {
			return r.URL.Path == "/health" && respStatus == http.StatusOK
		},
	})
}

func (s *Server) makeRequestHandler(ctx context.Context, svcs *rez.Services) (http.Handler, error) {
	auth, authErr := oidc.NewAuthSessionService(ctx, svcs.Organizations, svcs.Users)
	if authErr != nil {
		return nil, fmt.Errorf("oidc.NewAuthSessionService: %w", authErr)
	}

	r := chi.NewRouter()

	r.Get("/health", s.makeHealthCheckHandler())

	webhooks := chi.NewMux()
	for route, wh := range integrations.GetWebhookHandlers() {
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

		api := oapiv1.MakeApi(apiv1.NewHandler(svcs), oapiv1.MakeSecurityMiddleware(), oapiv1.MakeAPITelemetryMiddleware())
		ar.Mount(oapiv1.VersionPrefix, api.Adapter())
	})

	return r, nil
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
