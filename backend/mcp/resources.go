package mcp

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ReadResourceResult = mcp.ReadResourceResult
type ResourceContents = mcp.ResourceContents

type ResourcesHandler interface {
	ListActiveIncidents(ctx context.Context) ([]string, error)
	GetOncallShift(ctx context.Context, id uuid.UUID) ([]ResourceContents, error)
}

func addResources(s *server.MCPServer, h ResourcesHandler) {
	s.AddResource(getOncallShiftResource(h))
	s.AddResource(listActiveIncidentsResource(h))
}

func getOncallShiftResource(h ResourcesHandler) (mcp.Resource, server.ResourceHandlerFunc) {
	res := mcp.NewResource(
		"shift://{shift_id}",
		"Oncall Shift",
		mcp.WithResourceDescription("Information about a specific oncall shift"),
		mcp.WithMIMEType("application/json"),
	)
	handler := func(ctx context.Context, r mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		id, idErr := extractIdParam(r.Params.URI)
		if idErr != nil {
			return nil, idErr
		}
		return h.GetOncallShift(ctx, id)
	}
	return res, handler
}
func listActiveIncidentsResource(h ResourcesHandler) (mcp.Resource, server.ResourceHandlerFunc) {
	res := mcp.NewResource(
		"incidents://list",
		"List Incidents",
		mcp.WithResourceDescription("Provides a list of recent incidents"),
		mcp.WithMIMEType("application/json"),
	)
	handler := func(ctx context.Context, r mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		incs, incsErr := h.ListActiveIncidents(ctx)
		if incsErr != nil {
			return nil, incsErr
		}
		resp := make([]mcp.ResourceContents, len(incs))
		for i, inc := range incs {
			resp[i] = mcp.TextResourceContents{
				URI:      fmt.Sprintf("incidents://%s", uuid.New().String()),
				MIMEType: "text/markdown",
				Text:     inc,
			}
		}
		return resp, nil
	}
	return res, handler
}
