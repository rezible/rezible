package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/koding/websocketproxy"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/integrations"
	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	rez "github.com/rezible/rezible"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

type Config struct {
	Host           string               `koanf:"host"`
	Port           string               `koanf:"port"`
	DocumentsProxy DocumentsProxyConfig `koanf:"documents_proxy"`
}

type DocumentsProxyConfig struct {
	Enabled   bool   `koanf:"enabled"`
	ProxyHost string `koanf:"proxy_host"`
	serverUrl *url.URL
}

func loadConfig() (Config, error) {
	cfg := Config{
		Host: rez.Config.GetString("HOST", "0.0.0.0"),
		Port: rez.Config.GetString("PORT", "7002"),
		DocumentsProxy: DocumentsProxyConfig{
			Enabled:   false,
			ProxyHost: "localhost:7003",
		},
	}
	if cfgErr := rez.Config.Unmarshal("server.http", &cfg); cfgErr != nil {
		return cfg, fmt.Errorf("failed to unmarshal config: %w", cfgErr)
	}
	if cfg.DocumentsProxy.Enabled {
		proxyUrl, parseErr := url.Parse("ws://" + cfg.DocumentsProxy.ProxyHost)
		if parseErr != nil {
			return cfg, fmt.Errorf("failed to parse documents_proxy.proxy_host: %w", parseErr)
		}
		cfg.DocumentsProxy.serverUrl = proxyUrl
	}

	return cfg, nil
}

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

func NewServer(auth rez.AuthSessionService, v1h oapiv1.Handler) (*Server, error) {
	cfg, cfgErr := loadConfig()
	if cfgErr != nil {
		return nil, fmt.Errorf("config error: %w", cfgErr)
	}
	s := Server{cfg: cfg}

	s.router = chi.NewRouter()
	s.router.Use(s.loggerMiddleware)
	s.router.Use(middleware.Recoverer)

	basePath := rez.Config.BasePath()
	s.router.Mount(ensureSlashPrefix(basePath), http.StripPrefix(basePath, s.makeApiHandler(auth, v1h)))

	return &s, nil
}

const healthCheckPath = "/health"

func (s *Server) loggerMiddleware(next http.Handler) http.Handler {
	chiLogger := middleware.Logger(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != healthCheckPath {
			chiLogger.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (s *Server) makeApiHandler(auth rez.AuthSessionService, v1h oapiv1.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get(healthCheckPath, s.makeHealthCheckHandler())
	r.Mount(oapiv1.VersionPrefix, oapiv1.MakeApi(v1h, auth).Adapter())
	r.Mount("/auth", auth.AuthHandler())
	r.Mount("/webhooks", s.makeWebhooksHandler())
	if s.cfg.DocumentsProxy.Enabled {
		r.Handle("/documents", s.makeDocumentsProxyHandler(auth))
	}
	return r
}

func (s *Server) makeDocumentsProxyHandler(auth rez.AuthSessionService) http.Handler {
	headerKey := "X-Rez-Tenant-ID"
	setAuthContext := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookieVal := oapiv1.GetRequestAppCookieValue(r)
			if cookieVal == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			authCtx, authErr := auth.SetAuthSessionContext(r.Context(), cookieVal, "")
			if authErr != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			tenantId, tenantOk := access.GetTenantId(authCtx)
			if !tenantOk {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			r.Header.Set(headerKey, fmt.Sprintf("%d", tenantId))
			next.ServeHTTP(w, r.WithContext(authCtx))
		})
	}
	copyTenantIdHeaderFn := func(r *http.Request, h http.Header) {
		h.Set(headerKey, r.Header.Get(headerKey))
	}
	proxy := websocketproxy.NewProxy(s.cfg.DocumentsProxy.serverUrl)
	proxy.Director = copyTenantIdHeaderFn
	return chi.Chain(setAuthContext).Handler(proxy)
}

func (s *Server) makeHealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) makeWebhooksHandler() http.Handler {
	r := chi.NewMux()
	for route, wh := range integrations.GetWebhookHandlers() {
		r.Mount(ensureSlashPrefix(route), wh)
	}
	return r
}

func (s *Server) Start(baseCtx context.Context) error {
	s.httpServer = &http.Server{
		Addr:    net.JoinHostPort(s.cfg.Host, s.cfg.Port),
		Handler: s.router,
		BaseContext: func(l net.Listener) context.Context {
			return baseCtx
		},
	}

	log.Info().Msgf("HTTP server listening on %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server error: %w", err)
	}
	log.Info().Msgf("Stopped HTTP server")
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
		log.Error().Err(wErr).Msg("failed to write embedded docs body")
	}
}
