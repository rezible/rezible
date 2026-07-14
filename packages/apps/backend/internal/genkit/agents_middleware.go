package genkit

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type (
	AgentOutputToolResult struct {
		Status string `json:"status"`
	}

	outputWithWriteToolInfo interface {
		WriteToolInfo() (name string, description string)
	}
)

type agentOutputMiddleware[S rezai.SessionState, O rezai.AgentOutput] struct {
	partFn func(O) (*ai.Part, error)
}

func (a *agentOutputMiddleware[S, O]) Name() string {
	return "agent_output"
}

func (a *agentOutputMiddleware[S, O]) New(ctx context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		Tools: []ai.Tool{a.makeWriteOutputArtifactTool()},
	}, nil
}

func (a *agentOutputMiddleware[S, O]) makeWriteOutputArtifactTool() *aix.Tool[O, AgentOutputToolResult] {
	artifactName := "output"
	toolFunc := func(ctx context.Context, output O) (AgentOutputToolResult, error) {
		ss := aix.SessionFromContext[S](ctx)
		if ss == nil {
			return AgentOutputToolResult{Status: "Internal Error: no artifact store found in context (do not retry)"}, nil
		}
		part, partErr := a.partFn(output)
		if partErr != nil {
			return AgentOutputToolResult{Status: "failed to encode output: " + partErr.Error()}, nil
		}
		outputArtifact := &aix.Artifact{Name: artifactName, Parts: []*ai.Part{part}}
		for _, art := range ss.Artifacts() {
			if art.Name == artifactName {
				outputArtifact.Parts = append(art.Parts, outputArtifact.Parts...)
				outputArtifact.Metadata = art.Metadata
				break
			}
		}
		ss.AddArtifacts(outputArtifact)

		return AgentOutputToolResult{Status: "success"}, nil
	}
	toolName := "write_output"
	toolDesc := "Writes outputs of an agent run. For example a chat message response."
	var oe O
	if info, ok := any(oe).(outputWithWriteToolInfo); ok {
		toolName, toolDesc = info.WriteToolInfo()
	}
	return aix.NewTool(toolName, toolDesc, toolFunc)
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
