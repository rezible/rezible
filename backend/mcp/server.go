package mcp

import (
	"context"
	"net/http"

	"github.com/mark3labs/mcp-go/server"
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

func NewHTTPServer(h Handler, path string) http.Handler {
	ctxFn := func(ctx context.Context, r *http.Request) context.Context {
		return ctx
	}

	return server.NewStreamableHTTPServer(NewServer(h),
		server.WithEndpointPath(path),
		server.WithStateLess(true),
		server.WithHTTPContextFunc(ctxFn))
}
