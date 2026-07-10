package genkit

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"google.golang.org/genai"
)

var flashModel = googlegenai.ModelRef("googleai/gemini-flash-latest", &genai.GenerateContentConfig{
	ThinkingConfig: &genai.ThinkingConfig{ThinkingBudget: new(int32(0))},
})

func (s *AiService) getModel(name string) ai.ModelRef {
	return flashModel
}
