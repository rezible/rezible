package ai

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type (
	EvalScenarioDefinition struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		AgentName   string `json:"agentName"`
	}

	EvalScenarioSeed struct {
		Input            AgentInput
		SystemAnalysisID *uuid.UUID
	}

	EvalScenario interface {
		Definition() EvalScenarioDefinition
		Seed(context.Context, *ent.Client) (EvalScenarioSeed, error)
		Judge(context.Context, *ent.Client, *rez.AiAgentInvocationResult) ([]ai.Score, error)
	}
)
