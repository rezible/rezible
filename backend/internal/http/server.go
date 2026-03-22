package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/rezible/rezible/integrations"
	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	rez "github.com/rezible/rezible"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

type Server struct {
	router *chi.Mux

	httpServer *http.Server
}

func NewServer(auth rez.AuthService, v1h oapiv1.Handler) *Server {
	var s Server

	s.router = chi.NewRouter()
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	s.router.Mount(rez.Config.ApiPath(), s.makeApiHandler(auth, v1h))

	return &s
}

func (s *Server) makeApiHandler(auth rez.AuthService, v1h oapiv1.Handler) http.Handler {
	apiRouter := chi.NewRouter()

	apiRouter.Get("/health", s.healthCheckHandler)
	apiRouter.Mount(oapiv1.VersionPrefix, s.makeApiV1Handler(auth, v1h))
	apiRouter.Mount(rez.Config.WebhooksPath(), s.makeWebhooksHandler())

	return apiRouter
}

func (s *Server) makeApiV1Handler(auth rez.AuthService, v1h oapiv1.Handler) http.Handler {
	return oapiv1.MakeApi(v1h, auth).Adapter()
}

func (s *Server) makeWebhooksHandler() http.Handler {
	r := chi.NewMux()
	for route, wh := range integrations.GetWebhookHandlers() {
		r.Mount("/"+strings.TrimPrefix(route, "/"), wh)
	}
	return r
}

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) Start(baseCtx context.Context) error {
	addr := net.JoinHostPort(rez.Config.ListenHost(), rez.Config.ListenPort())
	s.httpServer = &http.Server{
		Addr:    addr,
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
