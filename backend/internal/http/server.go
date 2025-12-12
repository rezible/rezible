package http

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/mcp"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

type Server struct {
	auth rez.AuthService

	baseHandler http.Handler
	api         *chi.Mux

	httpServer *http.Server
}

func NewServer(auth rez.AuthService) *Server {
	s := Server{
		auth:        auth,
		baseHandler: http.NotFoundHandler(),
	}

	s.api = chi.NewMux()
	s.api.Mount(rez.Config.AuthRouteBase(), auth.AuthRouteHandler())
	s.api.Get("/health", s.healthCheckHandler)

	return &s
}

func (s *Server) commonMiddleware() chi.Middlewares {
	return chi.Chain(middleware.Logger)
}

func ensureSlashPrefix(s string) string {
	if strings.HasPrefix(s, "/") {
		return s
	}
	return "/" + s
}

func (s *Server) AddWebhookPathHandler(path string, handler http.Handler) {
	whPath := "/webhooks" + ensureSlashPrefix(path)
	s.api.Mount(whPath, http.StripPrefix(whPath, handler))
}

func (s *Server) MountMCP(h mcp.Handler) {
	mcpRouter := chi.Chain(s.auth.MCPServerMiddleware()).
		Handler(mcp.NewHTTPServer(h, "/mcp"))
	s.api.Mount("/mcp", mcpRouter)
}

func (s *Server) MountStaticFrontend(feFiles fs.FS) {
	s.baseHandler = makeEmbeddedFrontendFilesServer(feFiles)
}

func (s *Server) MountOpenApiV1(h oapiv1.Handler) {
	handler := oapiv1.MakeApi(h, rez.Config.ApiRouteBase(), oapiv1.MakeSecurityMiddleware(s.auth)).Adapter()
	oapiV1Router := s.commonMiddleware().Handler(handler)
	s.api.Mount("/v1", oapiV1Router)
}

func (s *Server) Start(baseCtx context.Context) error {
	r := chi.NewMux()
	r.Use(middleware.Recoverer)

	r.Mount(rez.Config.ApiRouteBase(), s.api)
	r.Handle("/*", s.baseHandler)

	host := rez.Config.GetStringOr("listen_host", "0.0.0.0")
	port := rez.Config.GetStringOr("listen_port", "8888")

	s.httpServer = &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: r,
	}

	s.httpServer.BaseContext = func(l net.Listener) context.Context {
		return baseCtx
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

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
