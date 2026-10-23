package genkit

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
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

func makeDefinedTool[I, O any](def rezai.ToolDefinition[I, O], toolFn func(context.Context, I) (O, error)) ai.Tool {
	return ai.NewTool(def.Name(), def.Description(), func(tc *ai.ToolContext, input I) (O, error) {
		return toolFn(tc, input)
	})
}
