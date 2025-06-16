package mcp

import (
	"context"
	"github.com/mark3labs/mcp-go/server"
	"net/http"
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

func NewStreamableHTTPServer(h Handler) *server.StreamableHTTPServer {
	ctxFn := func(ctx context.Context, r *http.Request) context.Context {
		return ctx
	}

	return server.NewStreamableHTTPServer(NewServer(h),
		server.WithEndpointPath("/mcp"),
		server.WithStateLess(true),
		server.WithHTTPContextFunc(ctxFn))
}
