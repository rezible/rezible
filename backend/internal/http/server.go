package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	oapi "github.com/rezible/rezible/openapi"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
)

type Server struct {
	httpServer *http.Server
}

func mount(r chi.Router, prefix string, h http.Handler) {
	r.Mount(prefix, http.StripPrefix(prefix, h))
}

func NewServer(addr string, pl rez.ProviderLoader, auth rez.AuthService, apiAdapter oapi.Adapter) (*Server, error) {
	s := Server{}

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)

	embeddedFeServer, feErr := makeEmbeddedFrontendServer()
	if feErr != nil {
		return nil, fmt.Errorf("failed to make embedded frontend server: %w", feErr)
	}

	frontendMiddleware := chi.Chain(
		makeAuthMiddleware(auth, true, []string{"/favicon.ico"}),
	)
	frontendHandler := frontendMiddleware.Handler(embeddedFeServer)

	apiMiddleware := chi.Chain(
		middleware.Logger,
		makeAuthMiddleware(auth, false, []string{"/openapi.json"}),
	)
	apiHandler := apiMiddleware.Handler(apiAdapter)

	webhookHandler := makeWebhookHandler(pl.HandleWebhookRequest)

	/* /api/ - API Routing Group */
	apiGroup := router.Group(func(r chi.Router) {
		/* /api/v1/ - OpenAPI Operations */
		mount(r, "/v1", apiHandler)

		/* /api/webhooks/ - Webhook routes */
		mount(r, "/webhooks", webhookHandler)

		/* /api/docs/ - OpenAPI Docs */
		r.Handle("/docs", makeApiDocsHandler())
	})
	mount(router, "/api", apiGroup)

	/* /auth/ - Auth Service Routing */
	router.Mount("/auth", auth.MakeAuthHandler())

	/* Serve static files for any other route */
	router.Handle("/*", frontendHandler)

	router.Get("/health", s.handleHealthcheck)

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

func (s *Server) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func makeWebhookHandler(providerHandler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerHandler.ServeHTTP(w, r)
	})
}

func makeAuthMiddleware(s rez.AuthService, redirect bool, skipPaths []string) func(http.Handler) http.Handler {
	authMw := s.MakeRequireAuthMiddleware(redirect)
	skipMap := make(map[string]struct{}, len(skipPaths))
	for _, path := range skipPaths {
		skipMap[path] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		withAuth := authMw(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, skip := skipMap[r.URL.Path]; skip {
				next.ServeHTTP(w, r)
				return
			}
			withAuth.ServeHTTP(w, r)
		})
	}
}
