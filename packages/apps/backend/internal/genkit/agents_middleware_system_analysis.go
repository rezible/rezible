package genkit

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	rez "github.com/rezible/rezible"
)

func WithSystemAnalysisMiddleware(as rez.AgentSessionService, sa rez.SystemAnalysisService, kg rez.KnowledgeGraphService) AgentMiddlewareConstructorFn {
	return func(AgentDetails) ai.Middleware {
		return newSystemAnalysisMiddleware(as, sa, kg)
	}
}

type systemAnalysisMiddleware struct {
	agentSessions rez.AgentSessionService
	analyses      rez.SystemAnalysisService
	knowledge     rez.KnowledgeGraphService
}

func newSystemAnalysisMiddleware(agentSessions rez.AgentSessionService, analyses rez.SystemAnalysisService, knowledge rez.KnowledgeGraphService) *systemAnalysisMiddleware {
	return &systemAnalysisMiddleware{agentSessions: agentSessions, analyses: analyses, knowledge: knowledge}
}

func (m *systemAnalysisMiddleware) Name() string {
	return "system_analysis"
}

func (m *systemAnalysisMiddleware) New(context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{Tools: []ai.Tool{}}, nil
}
