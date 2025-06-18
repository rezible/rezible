package mcp

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type PromptMessage = mcp.PromptMessage

type PromptsHandler interface {
	IncidentDebrief(context.Context, uuid.UUID) ([]mcp.PromptMessage, error)
}

func addPrompts(s *server.MCPServer, h PromptsHandler) {
	s.AddPrompt(makeIncidentDebriefPrompt(h))
}

type (
	IncidentDebriefPromptRequest struct {
		DebriefId uuid.UUID
	}
)

func makeIncidentDebriefPrompt(h PromptsHandler) (mcp.Prompt, server.PromptHandlerFunc) {
	debriefPrompt := mcp.NewPrompt("incident_debrief",
		mcp.WithPromptDescription("Discuss an incident with a user"),
		mcp.WithArgument("debrief_id",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("The debrief"),
		),
	)
	handler := func(ctx context.Context, r mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		debriefId, idErr := uuid.Parse(r.Params.Arguments["debrief_id"])
		if idErr != nil {
			return nil, fmt.Errorf("invalid id: %w", idErr)
		}
		msgs, msgErr := h.IncidentDebrief(ctx, debriefId)
		if msgErr != nil {
			return nil, msgErr
		}
		return &mcp.GetPromptResult{
			Description: "Incident debrief",
			Messages:    msgs,
		}, nil
	}
	return debriefPrompt, handler
}
