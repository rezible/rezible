package eino

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/mcp"
)

type MCPHandler struct {
	auth rez.AuthService
}

var _ mcp.Handler = &MCPHandler{}

func NewMCPHandler(auth rez.AuthService) *MCPHandler {
	return &MCPHandler{auth: auth}
}

func (m *MCPHandler) ListActiveIncidents(ctx context.Context) ([]mcp.ResourceContents, error) {
	incs := []mcp.ResourceContents{
		mcp.NewMarkdownResource("incidents://foo", "Example Incident"),
	}
	return incs, nil
}
