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
	httpServer *http.Server
}

func NewServer(
	addr string,
	auth rez.AuthSessionService,
	oapiHandler oapi.Handler,
	feFiles fs.FS,
	webhooksRouter http.Handler,
	mcpHandler mcp.Handler,
) *Server {
	var s Server

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)

	apiV1Router := chi.
		Chain(middleware.Logger).
		Handler(oapi.MakeApi(oapiHandler, "/api/v1", oapi.MakeSecurityMiddleware(auth)).Adapter())
	router.Mount("/api/v1", apiV1Router)

	router.Get("/api/docs", serveApiDocs)

	router.Mount("/api/webhooks", webhooksRouter)

	mcpRouter := chi.
		Chain(auth.MakeMCPServerAuthMiddleware()).
		Handler(mcp.NewHTTPServer(mcpHandler, "/mcp"))
	router.Mount("/mcp", mcpRouter)

	router.Mount("/auth", auth.MakeUserAuthHandler())
	router.Get("/health", makeHealthCheckHandler())

	frontendRouter := chi.
		Chain(auth.MakeFrontendAuthMiddleware()).
		Handler(makeEmbeddedFrontendFilesServer(feFiles))
	// Serve static files for any other route
	router.Handle("/*", frontendRouter)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return &s
}

func (s *Server) Start(ctx context.Context) error {
	s.httpServer.BaseContext = func(l net.Listener) context.Context {
		return ctx
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
