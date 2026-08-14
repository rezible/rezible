package genkit

import (
	"github.com/firebase/genkit/go/ai"
	gk "github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"google.golang.org/genai"
)

type ModelDefinition struct {
	Name string
	opts *ai.ModelOptions
	fn   ai.ModelFunc
}

func WithDefinedModel(def ModelDefinition) AiServiceOption {
	return AiServiceOption{
		kind: AiServiceOptionKindModel,
		optFn: func(s *AiService) error {
			gk.DefineModel(s.gk, def.Name, def.opts, def.fn)
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
