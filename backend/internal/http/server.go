package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/cors"
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
	router.Mount("/mcp", makeMCPHandler(mcpHandler, auth))

	o2p, o2pErr := mcp.NewOAuth2Provider("/oauth2")
	if o2pErr != nil {
		return nil, fmt.Errorf("failed to make mcp oauth2 provider: %w", o2pErr)
	}
	router.Mount("/oauth2", o2p.MakeHandler())
	wk := router.Group(func(r chi.Router) {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins: []string{"https://*", "http://*", "http://localhost:6274"},
			AllowedMethods: []string{"GET", "OPTIONS"},
			AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders: []string{"Link"},

			AllowCredentials: false,
			MaxAge:           300,
		}))
		r.Get("/oauth-authorization-server", o2p.AuthorizationServerMetadataDiscoveryHandler)
		r.Get("/oauth-protected-resource", o2p.ProtectedResourceMetadataDiscoveryHandler)
	})
	router.Mount("/.well-known", wk)

	router.Mount("/auth", auth.MakeUserAuthHandler())
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
	return chi.Chain().Handler(mcp.NewHTTPServer(h, auth))
}

func makeOApiHandler(h oapi.Handler, prefix string, auth rez.AuthSessionService) http.Handler {
	api := oapi.MakeApi(h, prefix, oapi.MakeSecurityMiddleware(auth))
	return chi.Chain(middleware.Logger).Handler(api.Adapter())
}

func makeHealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}
