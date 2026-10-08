package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/execution"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
)

type (
	Server struct {
		logger *slog.Logger

		server *http.Server

		listenerMu    sync.Mutex
		listenerReady atomic.Bool
	}

	UserAuthProvider interface {
		MakeAuthHandler() http.Handler
	}

	WebhookHandlers map[string]http.Handler

	HealthCheckFunc = func(context.Context) map[string]error
)

const (
	healthCheckPath    = "/health"
	readinessCheckPath = "/ready"

	// Webhook URLs may carry a secret in their path, so the request logger skips them and the webhooks
	// router logs each request without the path below the handler's name.
	webhooksPath = "/webhooks"
	// Webhook token issuance responses carry a secret webhook URL.
	webhookTokenPathSuffix = "/webhook-token"
)

func NewServer(
	cfg rez.Config,
	userAuth UserAuthProvider,
	v1Api oapiv1.API,
	webhooks WebhookHandlers,
	healthFn HealthCheckFunc,
) (*Server, error) {
	s := &Server{
		logger: slog.Default().WithGroup("http"),
	}

	router := chi.NewMux()

	router.Get(healthCheckPath, s.makeHealthCheckHandler(healthFn))
	router.Get(readinessCheckPath, s.makeReadyCheckHandler())

	router.Mount(webhooksPath, s.makeWebhooksRouter(webhooks))

	router.Mount("/auth", userAuth.MakeAuthHandler())

	router.Mount(oapiv1.VersionPrefix, v1Api.Adapter())

	s.server = s.makeServer(cfg, router)

	return s, nil
}

func (s *Server) makeServer(cfg rez.Config, r *chi.Mux) *http.Server {
	handler := chi.NewRouter()
	handler.Use(s.makeSetRootExecutionContextMiddleware())
	httpCfg := cfg.HttpServer
	handler.Use(s.makeRequestLoggerMiddleware(cfg.App.DebugMode, httpCfg.BasePath))
	handler.Mount(ensureSlashPrefix(httpCfg.BasePath), http.StripPrefix(httpCfg.BasePath, r))
	return &http.Server{
		Addr:    net.JoinHostPort(httpCfg.Host, httpCfg.Port),
		Handler: handler,
	}
}

func (s *Server) makeWebhooksRouter(webhooks WebhookHandlers) http.Handler {
	router := chi.NewMux()
	for prefix, wh := range webhooks {
		route := ensureSlashPrefix(prefix)
		slog.Debug("mounting webhook handler", "route", route)
		router.With(s.makeWebhookRequestLoggerMiddleware(webhooksPath+route)).Mount(route, wh)
	}
	return router
}

// makeWebhookRequestLoggerMiddleware logs each request to one webhook handler by the handler's path, never the
// rest of the request path. Handlers log their own deliveries. The request logger skips webhook requests, so
// this middleware also recovers a handler's panic and logs it here, without the panic value, which may carry
// the request path.
func (s *Server) makeWebhookRequestLoggerMiddleware(handlerPath string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				level := slog.LevelInfo
				attrs := []slog.Attr{
					slog.String("method", r.Method),
					slog.String("path", handlerPath),
				}
				recovered := recover()
				if recovered != nil {
					if ww.Status() == 0 {
						ww.WriteHeader(http.StatusInternalServerError)
					}
					level = slog.LevelError
					attrs = append(attrs,
						slog.String("panic", fmt.Sprintf("%T", recovered)),
						slog.String("stack", string(debug.Stack())),
					)
				}
				attrs = append(attrs,
					slog.Int("status", ww.Status()),
					slog.Duration("duration", time.Since(start)),
				)
				s.logger.LogAttrs(r.Context(), level, "webhook request", attrs...)
				if recovered == http.ErrAbortHandler {
					panic(recovered)
				}
			}()
			next.ServeHTTP(ww, r)
		})
	}
}

func ensureSlashPrefix(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}
	return s
}

func (s *Server) makeSetRootExecutionContextMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			execCtx := execution.NewRootContext(r.Context(), execution.KindAnonymous, execution.SourceHTTP)
			next.ServeHTTP(w, r.WithContext(execCtx))
		})
	}
}

func (s *Server) makeHealthCheckHandler(hc HealthCheckFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// The injector reports every service it has created, with a nil error for each healthy one.
		failures := make(map[string]error)
		for service, checkErr := range hc(r.Context()) {
			if checkErr != nil {
				failures[service] = checkErr
			}
		}
		if len(failures) > 0 {
			s.logger.WarnContext(r.Context(), "health check failed", "failures", failures)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) makeReadyCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.listenerReady.Load() {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) makeRequestLoggerMiddleware(concise bool, basePath string) func(http.Handler) http.Handler {
	logFormat := httplog.SchemaECS.Concise(concise)
	isDebugHeaderSet := func(r *http.Request) bool {
		return r.Header.Get("Debug") == "reveal-body-logs"
	}
	isSecretResponse := func(r *http.Request) bool {
		return strings.HasSuffix(r.URL.Path, webhookTokenPathSuffix)
	}

	skipPaths := mapset.NewThreadUnsafeSet(healthCheckPath, readinessCheckPath)
	// The logged path includes the base path.
	skipPathPrefix := basePath + webhooksPath + "/"
	return httplog.RequestLogger(s.logger, &httplog.Options{
		Level:         slog.LevelInfo,
		Schema:        logFormat,
		RecoverPanics: true,

		Skip: func(req *http.Request, respStatus int) bool {
			if skipPaths.Contains(req.URL.Path) || strings.HasPrefix(req.URL.Path, skipPathPrefix) {
				return true
			}
			return respStatus == 404 || respStatus == 405
		},

		LogRequestHeaders:  []string{"Origin"},
		LogResponseHeaders: []string{},

		LogRequestBody: isDebugHeaderSet,
		LogResponseBody: func(r *http.Request) bool {
			return isDebugHeaderSet(r) && !isSecretResponse(r)
		},
	})
}

func (s *Server) Run(ctx context.Context, ready chan<- struct{}) error {
	s.server.BaseContext = func(net.Listener) context.Context {
		return ctx
	}

	listener, listenerErr := s.makeListener()
	if listenerErr != nil {
		return fmt.Errorf("listener: %w", listenerErr)
	}
	close(ready)

	slog.Info("HTTP server listening", "addr", s.server.Addr)

	if serveErr := s.server.Serve(listener); !errors.Is(serveErr, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server: %w", serveErr)
	}
	return nil
}

func (s *Server) Handler() http.Handler {
	return s.server.Handler
}

func (s *Server) makeListener() (net.Listener, error) {
	s.listenerMu.Lock()
	defer s.listenerMu.Unlock()
	if s.listenerReady.Load() {
		return nil, fmt.Errorf("HTTP server is already started")
	}
	l, listenerErr := net.Listen("tcp", s.server.Addr)
	if listenerErr != nil {
		return nil, fmt.Errorf("listener: %w", listenerErr)
	}
	s.listenerReady.Store(true)
	return l, nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.listenerMu.Lock()
	defer s.listenerMu.Unlock()
	s.listenerReady.Store(false)
	if s.server != nil {
		slog.Info("HTTP server shutting down")
		if shutdownErr := s.server.Shutdown(ctx); shutdownErr != nil {
			return errors.Join(fmt.Errorf("shutdown HTTP server: %w", shutdownErr), s.server.Close())
		}
	}
	return nil
}

var (
	docsBodyScalar = []byte(`<!doctype html>
<html lang="en">
	<head>
		<title>API Reference</title>
		<meta charset="utf-8" />
		<meta
		name="viewport"
		content="width=device-width, initial-scale=1" />
	</head>
	<body>
		<script id="api-reference" data-url="/api/v1/openapi.json"></script>
		<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
	</body>
</html>`)

	docsBodyStoplight = []byte(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="referrer" content="same-origin" />
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no" />
    <title>API Dev Docs</title>
    <link href="https://unpkg.com/@stoplight/elements/styles.min.css" rel="stylesheet" />
    <script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
  </head>
  <body style="height: 100vh;">
    <elements-api
      apiDescriptionUrl="/api/v1/openapi.json"
      router="hash"
      layout="sidebar"
      tryItCredentialsPolicy="same-origin"
    />
  </body>
</html>`)
)

func serveApiDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if _, wErr := w.Write(docsBodyScalar); wErr != nil {
		slog.Error("failed to write embedded docs body", "error", wErr)
	}
}
