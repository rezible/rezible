package genkit

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	inv "github.com/rezible/rezible/ent/investigation"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type InvestigationAgent struct {
	investigations rez.InvestigationService
	analyses       rez.SystemAnalysisService
	knowledge      rez.KnowledgeGraphQueryService
}

func NewInvestigationAgent(
	investigations rez.InvestigationService,
	analyses rez.SystemAnalysisService,
	knowledge rez.KnowledgeGraphQueryService,
) *InvestigationAgent {
	return &InvestigationAgent{
		investigations: investigations,
		analyses:       analyses,
		knowledge:      knowledge,
	}
}

func (a *InvestigationAgent) agentDefinition() rezai.InvestigationAgentDefinition {
	return rezai.InvestigationAgent
}

func (a *InvestigationAgent) makeMiddleware() []ai.Middleware {
	return []ai.Middleware{
		&investigationCapabilitiesMiddleware{
			investigations: a.investigations,
			analyses:       a.analyses,
			knowledge:      a.knowledge,
		},
	}
}

func (a *InvestigationAgent) makeInitialTurnInput(_ context.Context, input rezai.InvestigationAgentSessionInput) (*rez.AiAgentTurnInput, error) {
	return &rez.AiAgentTurnInput{Message: ai.NewUserTextMessage(input.Query)}, nil
}

func (a *InvestigationAgent) getCustomState(context.Context, *ent.AgentSession) (*rezai.InvestigationAgentState, error) {
	return &rezai.InvestigationAgentState{}, nil
}

func (a *InvestigationAgent) transformState(ctx context.Context, state *aix.SessionState[rezai.InvestigationAgentState]) (*aix.SessionState[rezai.InvestigationAgentState], error) {
	return state, nil
}

func (a *InvestigationAgent) transformStreamChunk(ctx context.Context, chunk *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error) {
	return chunk, nil
}

type investigationCapabilitiesMiddleware struct {
	investigations rez.InvestigationService
	analyses       rez.SystemAnalysisService
	knowledge      rez.KnowledgeGraphQueryService
}

func (m *investigationCapabilitiesMiddleware) Name() string {
	return "investigation_capabilities"
}

func (m *investigationCapabilitiesMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	invCtx := getAgentInvocationContext(ctx)
	if invCtx == nil || invCtx.Session == nil || invCtx.Turn == nil {
		return nil, fmt.Errorf("investigation requires session and turn context")
	}
	if invCtx.Session.ID == uuid.Nil || invCtx.Turn.ID == uuid.Nil {
		return nil, fmt.Errorf("investigation session and turn IDs are required")
	}
	if invCtx.Turn.AgentSessionID != invCtx.Session.ID {
		return nil, fmt.Errorf("agent turn does not belong to supplied session")
	}

	current, lookupErr := m.investigations.LookupInvestigation(ctx, inv.AgentSessionID(invCtx.Session.ID))
	if ent.IsNotFound(lookupErr) {
		return nil, fmt.Errorf("investigation for agent session: %w", rez.ErrNotFound)
	}
	if lookupErr != nil {
		return nil, fmt.Errorf("lookup investigation by agent session: %w", lookupErr)
	}
	if current == nil || current.ID == uuid.Nil || current.AgentSessionID != invCtx.Session.ID || current.SystemAnalysisID == uuid.Nil {
		return nil, fmt.Errorf("investigation is missing its session association or analysis ID")
	}

	i := &investigationInvocation{
		investigations:  m.investigations,
		analyses:        m.analyses,
		knowledge:       m.knowledge,
		investigationID: current.ID,
		analysisID:      current.SystemAnalysisID,
		turnID:          invCtx.Turn.ID,
	}

	return i.makeHooks(ctx, invCtx)
}
