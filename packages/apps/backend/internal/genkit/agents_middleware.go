package genkit

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type agentOutputMiddleware[O rezai.AgentOutput] struct {
	outputTool *aix.Tool[O, AgentOutputToolResult]
}

func (a *agentOutputMiddleware[O]) Name() string {
	return "agent_output_artifacts"
}

func (a *agentOutputMiddleware[O]) New(ctx context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		Tools: []ai.Tool{a.outputTool},
	}, nil
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
			"Query the knowledge graph",
			func(ctx context.Context, in queryKnowledgeGraphToolInput) (queryKnowledgeGraphToolOutput, error) {
				return queryKnowledgeGraphToolOutput{}, nil
			},
		),
	}

	return &ai.Hooks{Tools: tools}, nil
}
