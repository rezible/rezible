package mcp

import (
	"context"
	"net/http"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type (
	TextContent = mcp.TextContent

	Server = server.MCPServer
	Client = client.Client
)

func NewInProcessClient(s *Server) (*Client, error) {
	return client.NewInProcessClient(s)
}

type Handler interface {
	ResourcesHandler
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

func NewServer(h Handler) *server.MCPServer {
	hooks := &server.Hooks{}

	s := server.NewMCPServer("Rezible MCP", "0.0.1",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(false, false),
		server.WithPromptCapabilities(true),
		server.WithRecovery(),
		server.WithHooks(hooks))

	addResources(s, h)

	return s
}

func addResources(s *server.MCPServer, h ResourcesHandler) {
	s.AddResource(ActiveIncidentsResource, makeActiveIncidentsResourceHandler(h))
}
