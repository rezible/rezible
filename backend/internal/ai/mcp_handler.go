package ai

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
	//TODO implement me
	panic("implement me")
}

func (m *MCPHandler) Calculate(ctx context.Context, request *mcp.CalculateRequest) (float64, error) {
	return 0, nil
}

func (m *MCPHandler) IncidentDebrief(ctx context.Context, uuid uuid.UUID) ([]mcp.PromptMessage, error) {
	//TODO implement me
	panic("implement me")
}
