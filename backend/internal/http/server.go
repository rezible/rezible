package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
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

func NewServer(auth rez.AuthService, v1h oapiv1.Handler) (*Server, error) {
	var s Server

	rootHandler, rootErr := s.makeRootHandler()
	if rootErr != nil {
		return nil, fmt.Errorf("root handler: %w", rootErr)
	}

	apiBasePath := rez.Config.BasePath()
	apiHandler, apiErr := s.makeApiHandler(auth, v1h, apiBasePath)
	if apiErr != nil {
		return nil, fmt.Errorf("api handler: %w", apiErr)
	}

	s.router = chi.NewRouter()
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	s.router.Mount(apiBasePath, apiHandler)
	s.router.NotFound(rootHandler.ServeHTTP)

	printRoutes("/", s.router.Routes())

	return &s, nil
}

func printRoutes(prefix string, routes []chi.Route) {
	for _, r := range routes {
		var handlers []string
		for method := range r.Handlers {
			handlers = append(handlers, method)
		}
		path, _ := strings.CutSuffix(r.Pattern, "*")
		route, _ := url.JoinPath(prefix, path)
		if r.SubRoutes != nil {
			printRoutes(route, r.SubRoutes.Routes())
		}
	}
}

func (s *Server) makeRootHandler() (http.Handler, error) {
	if rez.Config.ServeFrontend() {
		return makeEmbeddedFrontendFilesServer()
	}
	return s.makeNotFoundHandler(), nil
}

func (s *Server) makeApiHandler(auth rez.AuthService, v1h oapiv1.Handler, prefix string) (http.Handler, error) {
	whHandler, whErr := s.makeWebhooksHandler()
	if whErr != nil {
		return nil, fmt.Errorf("make webhooks handler: %w", whErr)
	}

	apiRouter := chi.NewRouter()
	apiRouter.Mount(rez.Config.WebhooksPath(), whHandler)
	apiRouter.Mount(oapiv1.VersionPrefix, s.makeV1ApiHandler(auth, v1h, prefix))
	// if rez.Config.ServeFrontend() {
	apiRouter.Mount(rez.Config.AuthPath(), auth.AuthRouteHandler())
	apiRouter.Get("/health", s.makeHealthCheckHandler())

	return apiRouter, nil
}

func (s *Server) commonMiddleware() chi.Middlewares {
	return chi.Chain(middleware.Logger)
}

func (s *Server) makeWebhooksHandler() (http.Handler, error) {
	r := chi.NewMux()
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("webhook handler not found")
		http.NotFound(w, r)
	})
	for route, wh := range integrations.GetWebhookHandlers() {
		if !strings.HasPrefix(route, "/") {
			route = "/" + route
		}
		log.Debug().
			Str("route", route).
			Msg("adding webhook handler")
		r.Mount(route, wh)
	}
	return r, nil
}

func (s *Server) makeV1ApiHandler(auth rez.AuthService, h oapiv1.Handler, prefix string) http.Handler {
	apiHandler := oapiv1.MakeApi(h, prefix, oapiv1.MakeSecurityMiddleware(auth))
	return s.commonMiddleware().Handler(apiHandler.Adapter())
}

func (s *Server) makeHealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) makeNotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
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
