package mcp

import (
	"context"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ReadResourceResult = mcp.ReadResourceResult
type ResourceContents = mcp.ResourceContents

type ResourcesHandler interface {
	GetOncallShift(ctx context.Context, id uuid.UUID) ([]ResourceContents, error)
}

func addResources(s *server.MCPServer, h ResourcesHandler) {
	s.AddResource(makeGetOncallShift(h))
}

func makeGetOncallShift(h ResourcesHandler) (mcp.Resource, server.ResourceHandlerFunc) {
	getOncallShiftResource := mcp.NewResource(
		"shift://{shift_id}",
		"Oncall Shift",
		mcp.WithResourceDescription("Information about a specific oncall shift"),
		mcp.WithMIMEType("application/json"),
	)
	return getOncallShiftResource, func(ctx context.Context, r mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		id, idErr := extractIdParam(r.Params.URI)
		if idErr != nil {
			return nil, idErr
		}
		return h.GetOncallShift(ctx, id)
	}
}
