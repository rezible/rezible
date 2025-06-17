package mcp

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ToolsHandler interface {
	Calculate(context.Context, *CalculateRequest) (float64, error)
}

func addTools(s *server.MCPServer, h ToolsHandler) {
	s.AddTool(makeCalculateTool(h))
}

type (
	CalculateRequest struct {
		Operation string  `json:"operation" validate:"required"`
		X         float64 `json:"x"`
		Y         float64 `json:"y"`
	}
)

func makeCalculateTool(h ToolsHandler) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("calculate",
		mcp.WithDescription("Perform arithmetic operations"),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Enum("add", "subtract", "multiply", "divide"),
			mcp.Description("The arithmetic operation to perform"),
		),
		mcp.WithNumber("x", mcp.Required(), mcp.Description("First number")),
		mcp.WithNumber("y", mcp.Required(), mcp.Description("Second number")),
	)
	handler := func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var calc CalculateRequest
		if bindErr := r.BindArguments(&calc); bindErr != nil {
			log.Debug().Err(bindErr).Msg("failed to bind arguments")
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		res, resErr := h.Calculate(ctx, &calc)
		if resErr != nil {
			return nil, resErr
		}
		return mcp.NewToolResultText(fmt.Sprintf("%.2f", res)), nil
	}
	return tool, handler
}
