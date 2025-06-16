package ai

import (
	"context"
	"github.com/google/uuid"
	"github.com/rezible/rezible/mcp"
)

type MCPHandler struct {
}

var _ mcp.Handler = &MCPHandler{}

func NewMCPHandler() *MCPHandler {
	return &MCPHandler{}
}

func (m *MCPHandler) GetOncallShift(ctx context.Context, id uuid.UUID) ([]mcp.ResourceContents, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MCPHandler) Calculate(ctx context.Context, request *mcp.CalculateRequest) (float64, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MCPHandler) IncidentDebrief(ctx context.Context, uuid uuid.UUID) ([]mcp.PromptMessage, error) {
	//TODO implement me
	panic("implement me")
}
