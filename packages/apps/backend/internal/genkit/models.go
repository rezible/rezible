package genkit

import (
	"github.com/firebase/genkit/go/ai"
	gk "github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"google.golang.org/genai"
)

type ModelDefinition[Config any] struct {
	Name string
	opts *ai.ModelOptions
	fn   ai.ModelActionFunc[Config]
}

func WithDefinedModel[Config any](def ModelDefinition[Config]) AiServiceOption {
	return AiServiceOption{
		kind: AiServiceOptionKindModel,
		optFn: func(s *AiService) error {
			gk.DefineModelAction(s.gk, def.Name, def.opts, def.fn)
			return nil
		},
	}
}

var flashModel = googlegenai.ModelRef("googleai/gemini-flash-latest", &genai.GenerateContentConfig{
	ThinkingConfig: &genai.ThinkingConfig{ThinkingLevel: genai.ThinkingLevelMinimal},
})

func (s *AiService) getDefaultModel() *ai.ModelRef {
	if s.cfg.Gemini.Enabled {
		return &flashModel
	}
	return nil
}
