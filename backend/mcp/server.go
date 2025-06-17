package mcp

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mark3labs/mcp-go/server"

	rez "github.com/rezible/rezible"
)

type Handler interface {
	ResourcesHandler
	ToolsHandler
	PromptsHandler
}

func NewServer(h Handler) *server.MCPServer {
	hooks := &server.Hooks{}

	s := server.NewMCPServer("Rezible MCP", "0.0.1",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
		server.WithPromptCapabilities(true),
		server.WithRecovery(),
		server.WithHooks(hooks))

	addResources(s, h)
	addTools(s, h)
	addPrompts(s, h)

	return s
}

func NewHTTPServer(h Handler, auth rez.AuthSessionService) http.Handler {
	authMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			fakeSess := &rez.AuthSession{
				ExpiresAt: time.Now().Add(time.Hour),
				UserId:    uuid.New(),
			}
			next.ServeHTTP(w, r.WithContext(auth.CreateSessionContext(r.Context(), fakeSess)))
		})
	}

	ctxFn := func(ctx context.Context, r *http.Request) context.Context {
		return ctx
	}

	srv := server.NewStreamableHTTPServer(NewServer(h),
		server.WithEndpointPath("/mcp"),
		server.WithStateLess(true),
		server.WithHTTPContextFunc(ctxFn))

	return chi.Chain(authMw).Handler(srv)
}
