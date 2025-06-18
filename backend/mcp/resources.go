package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type (
	ResourceContents = mcp.ResourceContents
	TextResource     = mcp.TextResourceContents

	ResourcesHandler interface {
		ListActiveIncidents(ctx context.Context) ([]ResourceContents, error)
		GetOncallShift(ctx context.Context, id uuid.UUID) (ResourceContents, error)
	}
)

func addResources(s *server.MCPServer, h ResourcesHandler) {
	s.AddResource(getOncallShiftResource(h))
	s.AddResource(listActiveIncidentsResource(h))
}

func NewMarkdownResource(uri string, content string) mcp.ResourceContents {
	return &TextResource{
		URI:      uri,
		MIMEType: "text/markdown",
		Text:     content,
	}
}

func wrapSingleResource(c mcp.ResourceContents, err error) ([]mcp.ResourceContents, error) {
	return []mcp.ResourceContents{c}, err
}

func extractIdParam(uri string) (uuid.UUID, error) {
	parts := strings.Split(uri, "://")
	if len(parts) < 2 {
		return uuid.Nil, fmt.Errorf("mcp: invalid URI: %s", uri)
	}
	return uuid.Parse(parts[1])
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
		return wrapSingleResource(h.GetOncallShift(ctx, id))
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
		return h.ListActiveIncidents(ctx)
	}
	return res, handler
}
