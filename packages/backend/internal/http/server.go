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
	WebhooksPath   string               `koanf:"webhooks_path"`
	DocumentsProxy DocumentsProxyConfig `koanf:"documents_proxy"`
}

type DocumentsProxyConfig struct {
	Enabled       bool   `koanf:"enabled"`
	ProxyPath     string `koanf:"proxy_path"`
	ServerAddress string `koanf:"server_address"`
	serverUrl     *url.URL
}

func loadConfig() (Config, error) {
	cfg := Config{
		Host:         "0.0.0.0",
		Port:         "7002",
		WebhooksPath: "/webhooks",
		DocumentsProxy: DocumentsProxyConfig{
			Enabled:       false,
			ProxyPath:     "/documents",
			ServerAddress: "ws://localhost:7003",
		},
	}
	if cfgErr := rez.Config.Unmarshal("server.http", &cfg); cfgErr != nil {
		return cfg, fmt.Errorf("failed to unmarshal config: %w", cfgErr)
	}
	if cfg.DocumentsProxy.Enabled {
		proxyUrl, parseErr := url.Parse(cfg.DocumentsProxy.ServerAddress)
		if parseErr != nil {
			return cfg, fmt.Errorf("failed to parse documents_proxy.proxy_path: %w", parseErr)
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

func NewServer(auth rez.AuthService, v1h oapiv1.Handler) (*Server, error) {
	cfg, cfgErr := loadConfig()
	if cfgErr != nil {
		return nil, fmt.Errorf("config error: %w", cfgErr)
	}
	s := Server{cfg: cfg}

	s.router = chi.NewRouter()
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Mount(rez.Config.ApiPath(), s.makeApiHandler(auth, v1h))

	return &s, nil
}

func ensureSlashPrefix(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}
	return s
}

func (s *Server) makeApiHandler(auth rez.AuthService, v1h oapiv1.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/health", s.makeHealthCheckHandler())
	r.Mount(oapiv1.VersionPrefix, s.makeApiV1Handler(auth, v1h))
	r.Mount(ensureSlashPrefix(s.cfg.WebhooksPath), s.makeWebhooksHandler())
	if s.cfg.DocumentsProxy.Enabled {
		r.Handle(s.cfg.DocumentsProxy.ProxyPath, s.makeDocumentsProxyHandler(auth))
	}
	return r
}

func (s *Server) makeDocumentsProxyHandler(auth rez.AuthService) http.Handler {
	headerKey := "X-Rez-Tenant-ID"
	setAuthContext := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx, authErr := auth.CreateAuthSessionContext(r.Context(), oapiv1.GetRequestAuthCookieToken(r))
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

func (s *Server) makeApiV1Handler(auth rez.AuthService, v1h oapiv1.Handler) http.Handler {
	return oapiv1.MakeApi(v1h, auth).Adapter()
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
