package genkit

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type agentRunnerMiddleware[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput] struct {
	runner agentRunner[I, S, O]
	tools  []ai.Tool
}

func newAgentRunnerMiddleware[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput](r agentRunner[I, S, O], reqTools []ai.ToolRef) (*agentRunnerMiddleware[I, S, O], error) {
	mw := &agentRunnerMiddleware[I, S, O]{runner: r}
	if toolsErr := mw.makeRequiredDynamicTools(reqTools); toolsErr != nil {
		return nil, fmt.Errorf("required dynamic tools: %w", toolsErr)
	}
	return mw, nil
}

func (m *agentRunnerMiddleware[I, S, O]) Name() string {
	return "rezible_agent_runner"
}

func (m *agentRunnerMiddleware[I, S, O]) New(ctx context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{Tools: m.tools}, nil
}

func (m *agentRunnerMiddleware[I, S, O]) makeRequiredDynamicTools(reqTools []ai.ToolRef) error {
	var tools []ai.Tool

	for _, ref := range reqTools {
		tool, toolErr := m.supplyDynamicRunnerTool(ref.Name())
		if toolErr != nil {
			return fmt.Errorf("required tool '%s': %w", ref.Name(), toolErr)
		}
		tools = append(tools, tool)
	}

	var o O
	if wo, supportsWriteArtifact := any(o).(rezai.AgentOutputWithWriteArtifactTool); supportsWriteArtifact {
		toolName, toolDesc := wo.WriteArtifactToolDefinition()
		m.tools = append(m.tools, m.makeWriteOutputArtifactTool(toolName, toolDesc))
	}

	return nil
}

func (m *agentRunnerMiddleware[I, S, O]) supplyDynamicRunnerTool(name string) (ai.Tool, error) {
	switch name {
	// TODO: dynamic tools
	default:
		return nil, fmt.Errorf("not implemented")
	}
}

func (m *agentRunnerMiddleware[I, S, O]) makeWriteOutputArtifactTool(name, desc string) *aix.Tool[O, rezai.WriteAgentOutputArtifactToolOutput] {
	writeArtifactFn := func(ctx context.Context, output O) error {
		as := aix.ArtifactStoreFromContext(ctx)
		if as == nil {
			return fmt.Errorf("no session found in context (do not retry)")
		}

		part, partErr := m.runner.agentDefinition().MakeOutputArtifactPart(output)
		if partErr != nil {
			return fmt.Errorf("encode output artifact: %w", partErr)
		}

		oa := &aix.Artifact{Name: "output", Parts: []*ai.Part{part}}
		for _, art := range as.Artifacts() {
			if art.Name == oa.Name {
				oa.Parts = append(art.Parts, oa.Parts...)
				oa.Metadata = art.Metadata
				break
			}
		}
		as.AddArtifacts(oa)

		return nil
	}
	return aix.NewTool(name, desc, func(ctx context.Context, output O) (rezai.WriteAgentOutputArtifactToolOutput, error) {
		status := "success"
		if writeErr := writeArtifactFn(ctx, output); writeErr != nil {
			status = "failed to write: " + writeErr.Error()
		}
		return rezai.WriteAgentOutputArtifactToolOutput{Status: status}, nil
	})
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
