package http

import (
	"context"
	"errors"
	"fmt"
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

func NewServer(addr string, auth rez.AuthSessionService, oapiHandler oapi.Handler, webhookHandler http.Handler) (*Server, error) {
	var s Server

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)

	frontendHandler, frontendErr := makeFrontendHandler(auth)
	if frontendErr != nil {
		return nil, fmt.Errorf("failed to make frontend handler: %w", frontendErr)
	}

	/* /api/ - API Routing Group */
	apiGroup := router.Group(func(r chi.Router) {
		/* /api/v1/ - OpenAPI Operations */
		mount(r, "/v1", makeApiHandler(oapiHandler, auth))

		/* /api/webhooks/ - Webhook routes */
		mount(r, "/webhooks", webhookHandler)

		/* /api/docs/ - OpenAPI Docs */
		r.Handle("/docs", makeApiDocsHandler())
	})
	mount(router, "/api", apiGroup)

	/* /auth/ - Auth Service Routing */
	router.Mount("/auth", auth.MakeUserAuthHandler())

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

func makeFrontendHandler(s rez.AuthSessionService) (http.Handler, error) {
	serveEmbeddedFiles, feErr := makeEmbeddedFrontendServer()
	if feErr != nil {
		return nil, fmt.Errorf("failed to make embedded frontend server: %w", feErr)
	}

	requireAuth := s.MakeFrontendAuthMiddleware()
	authMw := func(next http.Handler) http.Handler {
		withAuth := requireAuth(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/favicon.ico" {
				next.ServeHTTP(w, r)
			} else {
				withAuth.ServeHTTP(w, r)
			}
		})
	}

	return chi.Chain(authMw).Handler(serveEmbeddedFiles), nil
}

func makeApiHandler(oapiHandler oapi.Handler, auth rez.AuthSessionService) http.Handler {
	authMw := makeApiAuthMiddleware(auth)
	serveApi := oapi.MakeApi(oapiHandler, authMw).Adapter()
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

func getRequestTokenValue(r *http.Request) (string, error) {
	cookieToken, cookieErr := getRequestAuthSessionTokenCookie(r)
	if cookieErr != nil {
		return "", fmt.Errorf("error getting token from cookie: %w", cookieErr)
	} else if cookieToken != "" {
		return cookieToken, nil
	}

	bearerToken, bearerErr := getRequestAuthorizationBearerToken(r)
	if bearerErr != nil {
		return "", fmt.Errorf("error getting bearer token from authorization header: %w", bearerErr)
	} else if bearerToken != "" {
		return bearerToken, nil
	}

	return "", rez.ErrNoAuthSession
}
