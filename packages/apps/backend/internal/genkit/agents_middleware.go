package genkit

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type agentRunnerMiddleware[I rezai.AgentInput, S rezai.SessionState] struct {
	runner agentRunner[I, S]
	tools  []ai.Tool
}

func newAgentRunnerMiddleware[I rezai.AgentInput, S rezai.SessionState](r agentRunner[I, S], reqTools []ai.ToolRef) (*agentRunnerMiddleware[I, S], error) {
	mw := &agentRunnerMiddleware[I, S]{runner: r}
	if toolsErr := mw.makeRequiredDynamicTools(reqTools); toolsErr != nil {
		return nil, fmt.Errorf("required dynamic tools: %w", toolsErr)
	}
	return mw, nil
}

func (m *agentRunnerMiddleware[I, S]) Name() string {
	return "rezible_agent_runner"
}

func (m *agentRunnerMiddleware[I, S]) New(ctx context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{Tools: m.tools}, nil
}

func (m *agentRunnerMiddleware[I, S]) makeRequiredDynamicTools(reqTools []ai.ToolRef) error {
	var tools []ai.Tool

	for _, ref := range reqTools {
		tool, toolErr := m.supplyDynamicRunnerTool(ref.Name())
		if toolErr != nil {
			return fmt.Errorf("required tool '%s': %w", ref.Name(), toolErr)
		}
		tools = append(tools, tool)
	}

	return nil
}

func (m *agentRunnerMiddleware[I, S]) supplyDynamicRunnerTool(name string) (ai.Tool, error) {
	switch name {
	// TODO: dynamic tools
	default:
		return nil, fmt.Errorf("not implemented")
	}
}

type (
	knowledgeGraphMiddleware[S rezai.SessionState] struct {
		kg rez.KnowledgeGraphService
	}
	queryKnowledgeGraphToolInput  struct{}
	queryKnowledgeGraphToolOutput struct{}
)

func newKnowledgeGraphMiddleware[S rezai.SessionState](kg rez.KnowledgeGraphService) *knowledgeGraphMiddleware[S] {
	return &knowledgeGraphMiddleware[S]{kg: kg}
}

func (kg *knowledgeGraphMiddleware[S]) Name() string {
	return "knowledge_graph"
}

func (kg *knowledgeGraphMiddleware[S]) New(ctx context.Context) (*ai.Hooks, error) {
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
