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

func mount(r chi.Router, prefix string, h http.Handler) {
	r.Mount(prefix, http.StripPrefix(prefix, h))
}

func NewServer(addr string, auth rez.AuthSessionService, oapiHandler oapi.Handler, webhookHandler http.Handler, mcpHandler mcp.Handler) (*Server, error) {
	var s Server

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)

	frontendHandler, frontendErr := makeFrontendHandler(auth)
	if frontendErr != nil {
		return nil, fmt.Errorf("failed to make frontend handler: %w", frontendErr)
	}

	router.Mount("/api/v1", makeApiHandler("/api/v1", oapiHandler, auth))
	router.Handle("/api/docs", makeApiDocsHandler())
	router.Mount("/api/webhooks", webhookHandler)
	router.Mount("/api/mcp", makeMCPHandler(mcpHandler, auth))

	/* /auth/ - Auth Service Routing */
	router.Mount("/auth", auth.MakeUserAuthHandler())

	router.Mount("/.well-known", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: serve well known files (eg oauth2 configuration)
	}))

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

func makeFrontendHandler(auth rez.AuthSessionService) (http.Handler, error) {
	serveEmbeddedFiles, feErr := makeEmbeddedFrontendServer()
	if feErr != nil {
		return nil, fmt.Errorf("failed to make embedded frontend server: %w", feErr)
	}

	return chi.Chain(auth.MakeFrontendAuthMiddleware()).Handler(serveEmbeddedFiles), nil
}

func makeMCPHandler(mcpHandler mcp.Handler, auth rez.AuthSessionService) http.Handler {
	return chi.Chain(auth.MakeMCPAuthMiddleware()).Handler(mcp.NewStreamableHTTPServer(mcpHandler))
}

func makeApiHandler(prefix string, handler oapi.Handler, auth rez.AuthSessionService) http.Handler {
	authMw := makeApiAuthMiddleware(auth)
	serveApi := oapi.MakeApi(handler, prefix, authMw).Adapter()
	return chi.Chain(middleware.Logger).Handler(serveApi)
}

func makeApiAuthMiddleware(auth rez.AuthSessionService) func(oapi.Context, func(oapi.Context)) {
	return func(c oapi.Context, next func(oapi.Context)) {
		// var requireScopes []string
		security := c.Operation().Security
		explicitNoAuth := security != nil && len(security) == 0
		if explicitNoAuth {
			next(c)
			return
		}

		var sess *rez.AuthSession

		r, w := oapi.Unwrap(c)
		// TODO: check allowed token transports
		token, authErr := getRequestTokenValue(r)
		if token != "" {
			sess, authErr = auth.VerifySessionToken(token)
		}
		if authErr != nil {
			if writeErrRespErr := oapi.WriteAuthError(w, authErr); writeErrRespErr != nil {
				log.Error().Err(writeErrRespErr).Msg("failed to write auth error response")
			}
			return
		}

		authCtx := auth.CreateSessionContext(r.Context(), sess)
		next(oapi.WithContext(c, authCtx))
	}
}
