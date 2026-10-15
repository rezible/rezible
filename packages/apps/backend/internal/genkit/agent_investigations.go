package genkit

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	"github.com/rezible/rezible/pkg/errs"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	inv "github.com/rezible/rezible/ent/investigation"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type InvestigationAgent struct {
	deps investigationDependencies
}

// investigationDependencies are the services an investigation's tools read and write through.
type investigationDependencies struct {
	clock          rez.Clock
	investigations rez.InvestigationService
	outputs        rez.InvestigationOutputService
	knowledge      rez.KnowledgeGraphQueryService
	situations     rez.SituationService
	changes        rez.ChangesService
	integrations   rez.InstalledIntegrationGetter
	providerReads  rez.AgentProviderReadService
	proposals      rez.KnowledgeProposalService
}

func NewInvestigationAgent(
	clock rez.Clock,
	investigations rez.InvestigationService,
	outputs rez.InvestigationOutputService,
	knowledge rez.KnowledgeGraphQueryService,
	situations rez.SituationService,
	changes rez.ChangesService,
	integrations rez.InstalledIntegrationGetter,
	providerReads rez.AgentProviderReadService,
	proposals rez.KnowledgeProposalService,
) *InvestigationAgent {
	deps := investigationDependencies{
		clock:          clock,
		investigations: investigations,
		outputs:        outputs,
		knowledge:      knowledge,
		situations:     situations,
		changes:        changes,
		integrations:   integrations,
		providerReads:  providerReads,
		proposals:      proposals,
	}
	return &InvestigationAgent{deps: deps}
}

func (a *InvestigationAgent) agentDefinition() rezai.InvestigationAgentDefinition {
	return rezai.InvestigationAgent
}

func (a *InvestigationAgent) makeMiddleware() []ai.Middleware {
	return []ai.Middleware{
		&investigationCapabilitiesMiddleware{deps: a.deps},
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
	deps investigationDependencies
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

	current, lookupErr := m.deps.investigations.LookupInvestigation(ctx, inv.AgentSessionID(invCtx.Session.ID))
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			return nil, fmt.Errorf("investigation for agent session: %w", errs.ErrNotFound)
		}
		return nil, fmt.Errorf("lookup investigation by agent session: %w", lookupErr)
	}
	if current == nil || current.ID == uuid.Nil || current.AgentSessionID != invCtx.Session.ID {
		return nil, fmt.Errorf("investigation is missing its session association")
	}

	i := &investigationInvocation{
		investigationDependencies: m.deps,
		investigationID:           current.ID,
		turnID:                    invCtx.Turn.ID,
	}

	return i.makeHooks(ctx, invCtx)
}
