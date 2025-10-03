package http

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/mcp"
	oapi "github.com/rezible/rezible/openapi"
)

type Server struct {
	auth       rez.AuthService
	router     *chi.Mux
	httpServer *http.Server
	webhooks   *chi.Mux
}

func NewServer(addr string, auth rez.AuthService) *Server {
	var s Server

	s.auth = auth
	s.router = chi.NewRouter()
	s.router.Use(middleware.Recoverer)

	s.router.Mount("/auth", auth.UserAuthHandler())

	s.webhooks = chi.NewMux()
	s.mountWebhooks()

	s.router.Get("/health", makeHealthCheckHandler())

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	return &s
}

func (s *Server) commonMiddleware() chi.Middlewares {
	return chi.Chain(middleware.Logger)
}

func (s *Server) MountMCP(h mcp.Handler) {
	mcpRouter := chi.Chain(s.auth.MCPServerMiddleware()).
		Handler(mcp.NewHTTPServer(h, "/mcp"))
	s.router.Mount("/mcp", mcpRouter)
}

func (s *Server) MountDocuments(docs rez.DocumentsService) {
	docsApiRouter := s.commonMiddleware().Handler(docs.Handler())
	s.router.Mount("/api/documents", docsApiRouter)
}

func (s *Server) AddWebhookPathHandler(path string, handler http.Handler) {
	s.webhooks.Mount(path, handler)
}

func (s *Server) mountWebhooks() {
	webhooksRouter := s.commonMiddleware().Handler(http.StripPrefix("/api/webhooks", s.webhooks))
	s.router.Mount("/api/webhooks", webhooksRouter)
}

func (s *Server) MountStaticFrontend(feFiles fs.FS) {
	s.router.Handle("/*", makeEmbeddedFrontendFilesServer(feFiles))
}

func (s *Server) MountOpenApi(h oapi.Handler) {
	oapiServer := oapi.MakeApi(h, "/api/v1", oapi.MakeSecurityMiddleware(s.auth))
	oapiV1Router := chi.Chain(middleware.Logger).
		Handler(oapiServer.Adapter())
	s.router.Mount("/api/v1", oapiV1Router)
}

func (s *Server) Start(baseCtx context.Context) error {
	s.httpServer.BaseContext = func(l net.Listener) context.Context {
		return baseCtx
	}

	log.Info().Msgf("Serving on %s", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func makeHealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}
