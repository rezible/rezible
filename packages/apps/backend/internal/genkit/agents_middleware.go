package genkit

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type (
	AgentWriteOutputToolResult struct {
		Status string `json:"status"`
	}
)

type agentRunOutputWriter[S rezai.SessionState, O rezai.AgentOutput] struct {
	sessions rez.AiSessionStateService
}

func newAgentRunOutputWriter[S rezai.SessionState, O rezai.AgentOutput](sessions rez.AiSessionStateService) *agentRunOutputWriter[S, O] {
	return &agentRunOutputWriter[S, O]{sessions: sessions}
}

func (a *agentRunOutputWriter[S, O]) Name() string {
	return "output_writer"
}

func (a *agentRunOutputWriter[S, O]) New(ctx context.Context) (*ai.Hooks, error) {
	tools := []ai.Tool{
		aix.NewTool(
			"write_output",
			"Writes the results at the conclusion of your agent run. Use this to output deliverables.",
			func(ctx context.Context, in O) (AgentWriteOutputToolResult, error) {
				status := "Result saved successfully"
				if err := a.writeOutput(ctx, in); err != nil {
					status = fmt.Sprintf("Error writing result: %v", err)
				}
				return AgentWriteOutputToolResult{Status: status}, nil
			},
		),
	}

	return &ai.Hooks{Tools: tools}, nil
}

func (a *agentRunOutputWriter[S, O]) writeOutput(ctx context.Context, output O) error {
	sess := aix.SessionFromContext[S](ctx)
	if sess == nil {
		return fmt.Errorf("no session found in context")
	}
	runId, idErr := uuid.Parse(sess.SessionID())
	if idErr != nil {
		return fmt.Errorf("invalid session id: %v", sess.SessionID())
	}
	if validErr := output.Validate(); validErr != nil {
		return fmt.Errorf("validation error: %w", validErr)
	}
	if setErr := a.sessions.WriteAgentRunOutput(ctx, runId, output); setErr != nil {
		return fmt.Errorf("internal error saving the output. retrying will not succeed")
	}

	return nil
}

type (
	knowledgeGraphMiddleware struct {
		kg rez.KnowledgeGraphService
	}
	queryKnowledgeGraphToolInput  struct{}
	queryKnowledgeGraphToolOutput struct{}
)

func newKnowledgeGraphMiddleware(kg rez.KnowledgeGraphService) *knowledgeGraphMiddleware {
	return &knowledgeGraphMiddleware{kg: kg}
}

func (kg *knowledgeGraphMiddleware) Name() string {
	return "knowledge_graph"
}

func (kg *knowledgeGraphMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	tools := []ai.Tool{
		aix.NewTool(
			"query_knowledge_graph",
			"query the knowledge graph",
			func(ctx context.Context, in queryKnowledgeGraphToolInput) (queryKnowledgeGraphToolOutput, error) {
				return queryKnowledgeGraphToolOutput{}, nil
			},
		),
	}

	return &ai.Hooks{Tools: tools}, nil
}
