package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/koding/websocketproxy"
	"github.com/rezible/rezible/internal/http/oidc"
	"github.com/rezible/rezible/openapi"
	slogchi "github.com/samber/slog-chi"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/execution"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

type Server struct {
	cfg               rez.HttpServerConfig
	router            *chi.Mux
	logger            *slog.Logger
	httpServer        *http.Server
	documentsProxyUrl *url.URL
}

func NewServer(cfg rez.Config, ts rez.TelemetryService, authSess rez.AuthSessionService, oapiV1Handler oapiv1.Handler, webhookHandlers map[string]http.Handler) (*Server, error) {
	s := Server{
		cfg:    cfg.HttpServer,
		router: chi.NewRouter(),
		logger: slog.Default().WithGroup("http"),
	}

	if s.cfg.DocumentsProxy.Enabled {
		proxyUrl, parseErr := url.Parse("ws://" + s.cfg.DocumentsProxy.ProxyHost)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse documents_proxy.proxy_host: %w", parseErr)
		}
		s.documentsProxyUrl = proxyUrl
	}

	s.router.Use(s.makeSetRootExecutionContextMiddleware())
	s.router.Use(s.makeRequestLoggerMiddleware())

	apiHandler, apiErr := s.makeOpenApiAdapter(ts, oapiV1Handler)
	if apiErr != nil {
		return nil, fmt.Errorf("api adapter: %w", apiErr)
	}

	webhooksHandler := chi.NewMux()
	for route, wh := range webhookHandlers {
		webhooksHandler.Mount(ensureSlashPrefix(route), wh)
	}

	appCookie := oapiv1.NewAppCookie(cfg.App.FrontendApiPath)

	oidcAuthHandler, authErr := oidc.NewUserAuthHandler(cfg, authSess, appCookie)
	if authErr != nil {
		return nil, fmt.Errorf("user auth: %w", authErr)
	}

	ra := newRequestAuthenticator(authSess, appCookie)
	requestMw := chi.Middlewares{ra.middleware}

	handler := s.makeRequestHandler(requestMw, oidcAuthHandler, apiHandler, webhooksHandler)

	basePath := s.cfg.BasePath
	s.router.Mount(ensureSlashPrefix(basePath), http.StripPrefix(basePath, handler))

	return &s, nil
}

func (s *Server) makeOpenApiAdapter(ts rez.TelemetryService, v1h oapiv1.Handler) (openapi.Adapter, error) {
	verifyAuth := oapiv1.MakeAuthContextVerifier()
	userAuthMw := func(c openapi.Context, next func(openapi.Context)) {
		if ok := verifyAuth(c); ok {
			next(c)
		}
	}

	api := oapiv1.MakeApi(v1h, userAuthMw, oapiv1.MakeAPITelemetryMiddleware(ts))

	return api.Adapter(), nil
}

func (s *Server) makeSetRootExecutionContextMiddleware() func(http.Handler) http.Handler {
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

func (s *Server) makeRequestHandler(mw chi.Middlewares, auth http.Handler, api http.Handler, webhooks http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", s.makeHealthCheckHandler())
	r.Mount("/webhooks", webhooks)
	r.Mount("/auth", auth)

	// routes requiring auth session
	r.Group(func(ar chi.Router) {
		ar.Use(mw...)
		ar.Mount(oapiv1.VersionPrefix, api)
		if s.documentsProxyUrl != nil {
			ar.Handle("/documents", s.makeDocumentsProxyHandler(s.documentsProxyUrl))
		}
	})

	return r
}

func ensureSlashPrefix(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}
	return s
}

func (s *Server) makeDocumentsProxyHandler(serverUrl *url.URL) http.Handler {
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

	proxy := websocketproxy.NewProxy(serverUrl)
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

func (s *Server) Start(ctx context.Context) error {
	s.httpServer = &http.Server{
		Addr:    net.JoinHostPort(s.cfg.Host, s.cfg.Port),
		Handler: s.router,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	slog.Info("HTTP server listening", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server error: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("HTTP server shutting down")
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
