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
	AgentWriteResultToolInput[O rezai.AgentOutput] struct {
		Output O `json:"output"`
	}

	AgentWriteResultToolResult struct {
		Status string `json:"status"`
	}
)

type agentRunResultWriter[S rezai.SessionState, O rezai.AgentOutput] struct {
	sessions rez.AiSessionStateService
}

func newAgentRunResultWriter[S rezai.SessionState, O rezai.AgentOutput](sessions rez.AiSessionStateService) *agentRunResultWriter[S, O] {
	return &agentRunResultWriter[S, O]{sessions: sessions}
}

func (a *agentRunResultWriter[S, O]) Name() string {
	return "write_result"
}

func (a *agentRunResultWriter[S, O]) New(ctx context.Context) (*ai.Hooks, error) {
	tools := []ai.Tool{
		aix.NewTool(
			"write_result",
			"Writes the results at the conclusion of your agent run. Use this to output deliverables.",
			func(ctx context.Context, in AgentWriteResultToolInput[O]) (AgentWriteResultToolResult, error) {
				status := "Result saved successfully"
				if err := a.writeResult(ctx, in); err != nil {
					status = fmt.Sprintf("Error writing result: %v", err)
				}
				return AgentWriteResultToolResult{Status: status}, nil
			},
		),
	}

	return &ai.Hooks{Tools: tools}, nil
}

func (a *agentRunResultWriter[S, O]) writeResult(ctx context.Context, in AgentWriteResultToolInput[O]) error {
	sess := aix.SessionFromContext[S](ctx)
	if sess == nil {
		return fmt.Errorf("no session found in context")
	}
	runId, idErr := uuid.Parse(sess.SessionID())
	if idErr != nil {
		return fmt.Errorf("invalid session id: %v", sess.SessionID())
	}
	if validErr := in.Output.Validate(); validErr != nil {
		return fmt.Errorf("validation error: %w", validErr)
	}
	if setErr := a.sessions.SetAgentRunResultOutput(ctx, runId, in.Output); setErr != nil {
		return fmt.Errorf("save result error: %w", setErr)
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
