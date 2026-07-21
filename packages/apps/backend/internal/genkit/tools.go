package genkit

import (
	"context"

	genkitx "github.com/firebase/genkit/go/genkit/exp"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type ToolRunner[Input any, Output any] interface {
	Definition() rezai.ToolDefinition[Input, Output]
	ToolFunc(context.Context, Input) (Output, error)
}

func WithTool[I any, O any](t ToolRunner[I, O]) AiServiceOption {
	return AiServiceOption{
		kind: "tool",
		optFn: func(s *AiService) error {
			def := t.Definition()
			genkitx.DefineTool(s.gk, def.Name(), def.Description(), t.ToolFunc)
			return nil
		},
	}
}
