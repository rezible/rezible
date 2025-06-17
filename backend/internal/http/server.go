package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/rezible/rezible/mcp"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(
	addr string,
	auth rez.AuthSessionService,
	oapiHandler oapi.Handler,
	webhooksHandler http.Handler,
	mcpHandler mcp.Handler,
) (*Server, error) {
	var s Server

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)

	serveFrontendFiles, feFilesErr := makeEmbeddedFrontendFilesServer()
	if feFilesErr != nil {
		return nil, fmt.Errorf("failed to make embedded frontend server: %w", feFilesErr)
	}

	router.Mount("/api/v1", makeOApiHandler(oapiHandler, "/api/v1", auth))
	router.Handle("/api/docs", makeApiDocsHandler())
	router.Mount("/api/webhooks", webhooksHandler)
	router.Mount("/api/mcp", makeMCPHandler(mcpHandler, auth))
	router.Mount("/auth", auth.MakeUserAuthHandler())
	router.Mount("/.well-known", makeWellKnownHandler())
	router.Get("/health", makeHealthCheckHandler())

	// Serve static files for any other route
	router.Handle("/*", makeFrontendHandler(serveFrontendFiles, auth))

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return &s, nil
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

func makeFrontendHandler(feFilesHandler http.Handler, auth rez.AuthSessionService) http.Handler {
	return chi.Chain(auth.MakeFrontendAuthMiddleware()).Handler(feFilesHandler)
}

func makeMCPHandler(h mcp.Handler, auth rez.AuthSessionService) http.Handler {
	return chi.Chain(auth.MakeMCPAuthMiddleware()).Handler(mcp.NewStreamableHTTPServer(h))
}

func makeOApiHandler(h oapi.Handler, prefix string, auth rez.AuthSessionService) http.Handler {
	api := oapi.MakeApi(h, prefix, oapi.MakeSecurityMiddleware(auth))
	return chi.Chain(middleware.Logger).Handler(api.Adapter())
}

func makeWellKnownHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: serve well known files (eg oauth2 configuration)
		log.Debug().Str("path", r.URL.Path).Msg("Handling Well-Known Request")
		http.NotFound(w, r)
	})
}

func makeHealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}
