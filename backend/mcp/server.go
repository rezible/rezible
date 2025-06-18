package mcp

import (
	"context"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type (
	TextContent = mcp.TextContent
)

//const (
//	RoleUser      = mcp.RoleUser
//	RoleAssistant = mcp.RoleAssistant
//)

type Handler interface {
	ResourcesHandler
	ToolsHandler
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
	addTools(s, h)
	// addPrompts(s, h)

	return s
}

func addResources(s *server.MCPServer, h ResourcesHandler) {
	s.AddResourceTemplate(OncallShiftResource, makeOncallShiftResourceHandler(h))
	s.AddResource(ActiveIncidentsResource, makeActiveIncidentsResourceHandler(h))
}

func addTools(s *server.MCPServer, h ToolsHandler) {
	s.AddTool(CalculateTool, calculateToolHandler(h))
}
