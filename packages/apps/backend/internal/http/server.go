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
	"github.com/go-chi/httplog/v3"

	"github.com/koding/websocketproxy"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/http/oidc"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/openapi"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
)

type (
	Server struct {
		cfg    rez.HttpServerConfig
		router *chi.Mux
		logger *slog.Logger
	}
	WebhookHandlers map[string]http.Handler
)

func NewServer(cfg rez.Config, ts rez.TelemetryService, sess rez.AuthSessionService, oapiV1Handler oapiv1.Handler, webhooks WebhookHandlers) (*Server, error) {
	s := &Server{
		cfg:    cfg.HttpServer,
		logger: slog.Default().WithGroup("http"),
	}

	s.router = chi.NewRouter()
	s.router.Use(s.makeSetRootExecutionContextMiddleware())
	s.router.Use(s.makeRequestLoggerMiddleware(cfg.App.DebugMode))

	var documentsProxyUrl *url.URL
	if cfg.Documents.Proxy.Enabled {
		proxyUrl, parseErr := url.Parse("ws://" + cfg.Documents.Proxy.Host)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse documents_proxy.proxy_host: %w", parseErr)
		}
		documentsProxyUrl = proxyUrl
	}

	handler := chi.NewRouter()

	handler.Get("/health", s.makeHealthCheckHandler())

	webhooksHandler := chi.NewMux()
	for prefix, wh := range webhooks {
		route := ensureSlashPrefix(prefix)
		slog.Debug("mounting webhook handler", "route", route)
		webhooksHandler.Mount(route, wh)
	}
	handler.Mount("/webhooks", webhooksHandler)

	asc := newAppAuthSessionCookie(cfg.App.FrontendApiPath)
	oidcAuthHandler, authErr := oidc.NewUserAuthHandler(cfg, sess, asc)
	if authErr != nil {
		return nil, fmt.Errorf("user auth: %w", authErr)
	}
	handler.Mount("/auth", oidcAuthHandler)

	rv := newRequestAuthValidator(sess, asc)
	if cfg.HttpServer.Auth.EnableDevSkipMode {
		slog.Warn("enabling development session auth override")
		rv.devSessionOverride = true
	}
	// api routes with auth check
	handler.Group(func(ar chi.Router) {
		ar.Use(rv.AuthSessionMiddleware)

		ar.Mount(oapiv1.VersionPrefix, s.makeOpenApiHandler(ts, oapiV1Handler))

		if documentsProxyUrl != nil {
			ar.Handle("/documents", s.makeDocumentsProxyHandler(documentsProxyUrl))
		}
	})

	s.router.Mount(ensureSlashPrefix(s.cfg.BasePath), http.StripPrefix(s.cfg.BasePath, handler))

	return s, nil
}

func authScopesSatisfied(authScopes []string, secOpts oapiv1.SecurityMethodOptions) bool {
	authParts := make(map[string][]string)
	for _, scope := range authScopes {
		parts := strings.Split(scope, ":")
		if len(parts) == 2 || len(parts) == 3 {
			authParts[parts[0]] = parts[1:]
		} else {
			slog.Warn("invalid auth scope", "scope", scope)
		}
	}
	for _, opt := range secOpts {
		for method, scopes := range opt {
			slog.Debug("check api method scopes", "method", method, "scopes", scopes)
			for _, scope := range scopes {
				methodParts := strings.Split(scope, ":")
				if len(methodParts) != 2 && len(methodParts) != 3 {
					slog.Warn("invalid api security method scope",
						"method", method, "scope", scope)
					continue
				}
				subParts, ok := authParts[methodParts[0]]
				if !ok {
					continue
				}
				// TODO: check subParts
				slog.Debug("check scope sub parts", "subParts", subParts)
				return true
			}
		}
	}
	return false
}

func (s *Server) makeOpenApiHandler(ts rez.TelemetryService, v1h oapiv1.Handler) openapi.Adapter {
	checkMethodOptionsFn := func(ctx context.Context, secOpts oapiv1.SecurityMethodOptions) error {
		ec := execution.GetContext(ctx)

		if ec.IsAnonymous() {
			return rez.ErrAuthSessionMissing
		}

		if len(ec.Auth.Scopes) > 0 {
			if !authScopesSatisfied(ec.Auth.Scopes, secOpts) {
				return rez.ErrAuthSessionInvalid
			}
		}

		return nil
	}

	api := oapiv1.MakeApi(v1h,
		oapiv1.MakeRequestMethodSecurityMiddleware(checkMethodOptionsFn),
		oapiv1.MakeAPITelemetryMiddleware(ts))
	return api.Adapter()
}

func (s *Server) makeSetRootExecutionContextMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			execCtx := execution.NewRootContext(r.Context(), execution.KindAnonymous, execution.SourceHTTP)
			next.ServeHTTP(w, r.WithContext(execCtx))
		})
	}
}

func (s *Server) makeRequestLoggerMiddleware(concise bool) func(http.Handler) http.Handler {
	logFormat := httplog.SchemaECS.Concise(concise)
	isDebugHeaderSet := func(r *http.Request) bool {
		return r.Header.Get("Debug") == "reveal-body-logs"
	}

	return httplog.RequestLogger(s.logger, &httplog.Options{
		Level:         slog.LevelInfo,
		Schema:        logFormat,
		RecoverPanics: true,

		// Optionally, filter out some request logs.
		Skip: func(req *http.Request, respStatus int) bool {
			return respStatus == 404 || respStatus == 405
		},

		// Optionally, log selected request/response headers explicitly.
		LogRequestHeaders:  []string{"Origin"},
		LogResponseHeaders: []string{},

		// Optionally, enable logging of request/response body based on custom conditions.
		// Useful for debugging payload issues in development.
		LogRequestBody:  isDebugHeaderSet,
		LogResponseBody: isDebugHeaderSet,
	})
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
			if exec.IsAnonymous() || !tenantOk {
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

func (s *Server) Lifecycle() *rez.ServiceLifecycle {
	server := &http.Server{
		Addr:    net.JoinHostPort(s.cfg.Host, s.cfg.Port),
		Handler: s.router,
	}

	runFn := func(ctx context.Context) error {
		server.BaseContext = func(net.Listener) context.Context {
			return ctx
		}
		slog.Info("HTTP server listening", "addr", server.Addr)

		if srvErr := server.ListenAndServe(); !errors.Is(srvErr, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server: %w", srvErr)
		}
		return nil
	}
	stopFn := func(ctx context.Context) error {
		slog.Info("HTTP server shutting down")
		if shutdownErr := server.Shutdown(ctx); shutdownErr != nil {
			return errors.Join(fmt.Errorf("shutdown HTTP server: %w", shutdownErr), server.Close())
		}
		return nil
	}

	return &rez.ServiceLifecycle{
		StartFns: []rez.LifecycleFunc{runFn},
		StopFn:   stopFn,
	}
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
