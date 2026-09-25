package genkit

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/genkit"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
	rezai "github.com/rezible/rezible/pkg/ai"
)

func (r *AiRuntime) getRegisteredTools(refs []ai.ToolRef) ([]ai.ToolRef, []ai.ToolRef) {
	toolMap := make(map[string]ai.ToolRef)
	for _, ref := range genkit.ListTools(r.gk) {
		toolMap[ref.Name()] = ref
	}
	var registered []ai.ToolRef
	var missing []ai.ToolRef
	for _, t := range refs {
		if ref, ok := toolMap[t.Name()]; ok {
			registered = append(registered, ref)
		} else {
			missing = append(missing, t)
		}
	}
	return registered, missing
}

type ToolRunner[Input any, Output any] interface {
	Definition() rezai.ToolDefinition[Input, Output]
	ToolFunc(context.Context, Input) (Output, error)
}

func makeDefinedTool[I, O any](def rezai.ToolDefinition[I, O], toolFn aix.ToolFunc[I, O]) ai.Tool {
	return aix.NewTool(def.Name(), def.Description(), toolFn)
}

func WithDefinedTool[I any, O any](t ToolRunner[I, O]) AiRuntimeOption {
	return AiRuntimeOption{
		kind: AiRuntimeOptionKindTool,
		runtimeFn: func(s *AiRuntime) error {
			def := t.Definition()
			genkitx.DefineTool(s.gk, def.Name(), def.Description(), t.ToolFunc)
			return nil
		},
	}
}
