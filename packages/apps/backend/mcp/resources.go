package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type (
	ResourceContents     = mcp.ResourceContents
	TextResourceContents = mcp.TextResourceContents

	ResourcesHandler interface {
		ListActiveIncidents(ctx context.Context) ([]ResourceContents, error)
	}
)

func NewMarkdownResource(uri string, content string) mcp.ResourceContents {
	return &TextResourceContents{
		URI:      uri,
		MIMEType: "text/markdown",
		Text:     content,
	}
}

var ActiveIncidentsResource = mcp.NewResource(
	"incidents://active",
	"List Incidents",
	mcp.WithResourceDescription("Provides a list of recent incidents"),
	mcp.WithMIMEType("application/json"),
)

func makeActiveIncidentsResourceHandler(h ResourcesHandler) server.ResourceHandlerFunc {
	return func(ctx context.Context, r mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return h.ListActiveIncidents(ctx)
	}
}
