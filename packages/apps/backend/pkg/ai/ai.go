package ai

import (
	"embed"

	"github.com/firebase/genkit/go/ai"
)

type (
	Message = ai.Message
)

//go:embed prompts
var PromptsDir embed.FS
