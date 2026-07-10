package genkit

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/firebase/genkit/go/ai"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
	rez "github.com/rezible/rezible"
)

type Tool[Input any, Output any] interface {
	Name() string
	Description() string
	ToolFunc(context.Context, Input) (Output, error)
	ToolOpts() []ai.ToolOption
}

func WithTool[Input any, Output any](t Tool[Input, Output]) AiServiceOption {
	return func(s *AiService) error {
		ref := genkitx.DefineTool(s.gk, t.Name(), t.Description(), t.ToolFunc, t.ToolOpts()...)
		s.toolRefs = append(s.toolRefs, ref)
		return nil
	}
}

func (s *AiService) getRequiredToolRefs(names []string) ([]ai.ToolRef, error) {
	var refs []ai.ToolRef
	nameSet := mapset.NewSet(names...)
	for _, tool := range s.toolRefs {
		if !nameSet.Contains(tool.Name()) {
			return nil, fmt.Errorf("tool %s not found in tools %v", tool.Name(), names)
		}
		refs = append(refs, tool)
	}
	return refs, nil
}

type KnowledgeGraphTool struct {
	kg rez.KnowledgeGraphService
}

func NewKnowledgeGraphTool(kg rez.KnowledgeGraphService) *KnowledgeGraphTool {
	return &KnowledgeGraphTool{kg: kg}
}

func (kg *KnowledgeGraphTool) Name() string {
	return "query_knowledge_graph"
}

func (kg *KnowledgeGraphTool) Description() string {
	return ""
}

func (kg *KnowledgeGraphTool) ToolFunc(ctx context.Context, inp string) (string, error) {
	return "", nil
}

func (kg *KnowledgeGraphTool) ToolOpts() []ai.ToolOption {
	return nil
}
