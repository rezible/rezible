package eino

import (
	"context"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/mcp"
)

type MCPHandler struct {
	auth rez.AuthSessionService
}

var _ mcp.Handler = &MCPHandler{}

func NewMCPHandler(auth rez.AuthSessionService) *MCPHandler {
	return &MCPHandler{auth: auth}
}

func (m *MCPHandler) ListActiveIncidents(ctx context.Context) ([]mcp.ResourceContents, error) {
	incs := []mcp.ResourceContents{
		mcp.NewMarkdownResource("incidents://foo", "Example Incident"),
	}
	return incs, nil
}

func (m *MCPHandler) GetOncallShift(ctx context.Context, id uuid.UUID) (mcp.ResourceContents, error) {
	shiftRes := mcp.NewMarkdownResource("oncall_shifts://"+id.String(), "Example Shift")
	return shiftRes, nil
}
