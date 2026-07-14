package genkit

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/firebase/genkit/go/ai"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
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
