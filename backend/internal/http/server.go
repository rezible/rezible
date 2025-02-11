package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type Server struct {
	srv           *http.Server
	webhookRouter *chi.Mux
}

func (s *Server) Start(ctx context.Context) error {
	s.srv.BaseContext = func(l net.Listener) context.Context {
		return ctx
	}

	log.Info().Msgf("Serving on %s", s.srv.Addr)

	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}

func NewServer(addr string, pl rez.ProviderLoader, auth rez.AuthService, apiAdapter oapi.Adapter) (*Server, error) {
	var server Server

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	embeddedFeServer, feErr := makeEmbeddedFrontendServer()
	if feErr != nil {
		return nil, fmt.Errorf("failed to make embedded frontend server: %w", feErr)
	}

	frontendAuthMw := makeAuthMiddleware(auth, true, []string{"/favicon.ico"})
	frontendHandler := chi.Chain(
		frontendAuthMw,
	).Handler(embeddedFeServer)

	apiAuthMw := makeAuthMiddleware(auth, false, []string{"/openapi.json"})
	apiHandler := chi.Chain(
		middleware.Logger,
		apiAuthMw,
	).Handler(apiAdapter)

	webhookHandler := makeWebhookHandler(http.HandlerFunc(pl.HandleWebhookRequest))

	/* /api/ - API Routing Group */
	mount(router, "/api", router.Group(func(r chi.Router) {
		/* /api/v1/ - OpenAPI Operations */
		mount(r, "/v1", apiHandler)

		/* /api/webhooks/ - Webhook routes */
		mount(r, "/webhooks", webhookHandler)

		/* /api/docs/ - OpenAPI Docs */
		r.Handle("/docs", makeApiDocsHandler())
	}))

	/* /auth/ - Auth Service Routing */
	router.Mount("/auth", auth.MakeAuthHandler())

	/* Serve static files for any other route */
	router.Handle("/*", frontendHandler)

	server.srv = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return &server, nil
}

func mount(r chi.Router, prefix string, h http.Handler) {
	r.Mount(prefix, http.StripPrefix(prefix, h))
}

func makeWebhookHandler(providerHandler http.Handler) http.Handler {
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
