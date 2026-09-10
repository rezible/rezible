package genkit

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	siti "github.com/rezible/rezible/ent/situationinvestigation"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type InvestigationAgent struct {
	situations rez.SituationService
	analyses   rez.SystemAnalysisService
	knowledge  rez.KnowledgeGraphService
}

func NewInvestigationAgent(situations rez.SituationService, analyses rez.SystemAnalysisService, knowledge rez.KnowledgeGraphService) *InvestigationAgent {
	return &InvestigationAgent{situations: situations, analyses: analyses, knowledge: knowledge}
}

func (a *InvestigationAgent) agentDefinition() rezai.InvestigationAgentDefinition {
	return rezai.InvestigationAgent
}

func (a *InvestigationAgent) makeMiddleware() []ai.Middleware {
	resolveInvocationSystemAnalysisID := func(ctx context.Context) (uuid.UUID, error) {
		inv, invErr := a.resolveInvocationSituationInvestigation(ctx)
		if invErr != nil {
			return uuid.Nil, invErr
		}
		return inv.SystemAnalysisID, nil
	}
	return []ai.Middleware{
		newSystemAnalysisMiddleware(a.analyses, a.knowledge, resolveInvocationSystemAnalysisID),
		&situationInvestigationReportMiddleware{situations: a.situations},
	}
}

func (a *InvestigationAgent) makeInitialTurnInput(ctx context.Context, input rezai.InvestigationAgentSessionInput) (*rez.AiAgentTurnInput, error) {
	return &rez.AiAgentTurnInput{
		Message: ai.NewUserTextMessage("Investigate this situation."),
	}, nil
}

func (a *InvestigationAgent) resolveInvocationSituationInvestigation(ctx context.Context) (*ent.SituationInvestigation, error) {
	invCtx := getAgentInvocationContext(ctx)
	if invCtx == nil {
		return nil, fmt.Errorf("agent session context does not exist")
	}
	inv, lookupInvErr := a.situations.LookupSituationInvestigation(ctx, siti.AgentSessionID(invCtx.Session.ID))
	if lookupInvErr != nil || inv == nil {
		return nil, fmt.Errorf("get investigation for agent session: %w", lookupInvErr)
	}
	return inv, nil
}

func (a *InvestigationAgent) updateInitialTurnMessage(ctx context.Context, input rezai.InvestigationAgentSessionInput) (string, error) {
	inv, invErr := a.resolveInvocationSituationInvestigation(ctx)
	if invErr != nil {
		return "", fmt.Errorf("get investigation: %w", invErr)
	}

	sit, situationErr := inv.Edges.SituationOrErr()
	if situationErr != nil {
		return "", fmt.Errorf("situation: %w", situationErr)
	}

	return fmt.Sprintf(`Title: %s
Summary: %s
Status: %s
Opened at: %s`, sit.Title, sit.Summary, sit.Status, sit.OpenedAt.Format("2006-01-02T15:04:05Z07:00")), nil
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

type situationInvestigationReportMiddleware struct {
	situations rez.SituationService
}

func (m *situationInvestigationReportMiddleware) Name() string {
	return "situation_investigation_report"
}

func (m *situationInvestigationReportMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	invCtx := getAgentInvocationContext(ctx)
	if invCtx == nil {
		return nil, fmt.Errorf("agent session context does not exist")
	}

	inv, lookupInvErr := m.situations.LookupSituationInvestigation(ctx, siti.AgentSessionID(invCtx.Session.ID))
	if lookupInvErr != nil || inv == nil {
		return nil, fmt.Errorf("get investigation for agent session: %w", lookupInvErr)
	}

	hooks := &ai.Hooks{}

	requested := inv.RequestedTurnID != nil && *inv.RequestedTurnID == invCtx.Turn.ID
	if !requested {
		hooks.WrapGenerate = makeSystemTextInjectorFn("situation_report", "This is a conversational follow-up. The existing investigation report is: "+inv.Report.Text)
	} else {
		hooks.Tools = append(hooks.Tools, makeDefinedTool(rezai.SaveSituationInvestigationReportTool, m.updateReportToolFunc))
		hooks.WrapGenerate = m.makeGenerateWrapper(invCtx.Turn.ID, inv.SituationID, inv.ID)
	}
	return hooks, nil
}

func (m *situationInvestigationReportMiddleware) makeGenerateWrapper(turnId uuid.UUID, situationId uuid.UUID, investigationID uuid.UUID) wrapGenerateFn {
	return func(ctx context.Context, params *ai.GenerateParams, next ai.GenerateNext) (*ai.ModelResponse, error) {
		response, respErr := next(ctx, params)
		if respErr != nil {
			return nil, respErr
		}
		if response != nil && len(response.ToolRequests()) == 0 {
			current, getInvErr := m.situations.GetInvestigationForSituation(ctx, investigationID)
			if getInvErr != nil {
				return nil, getInvErr
			}
			if current.RequestedTurnID == nil || *current.RequestedTurnID != turnId || current.CompletedRevision < current.RequestedRevision {
				return nil, fmt.Errorf("investigation finished without an accepted report")
			}
		}
		return response, nil
	}
}

func (m *situationInvestigationReportMiddleware) updateReportToolFunc(ctx context.Context, input rezai.SaveSituationInvestigationReportToolInput) (*rezai.SaveSituationInvestigationReportToolOutput, error) {
	invCtx := getAgentInvocationContext(ctx)
	if invCtx == nil || invCtx.Turn == nil {
		return nil, fmt.Errorf("missing agent invocation context")
	}

	params := rez.SetSituationInvestigationReportParams{
		AgentTurnID: invCtx.Turn.ID,
		Report:      input.Report,
		Assessments: input.Assessments,
	}
	if _, setReportErr := m.situations.SetSituationInvestigationReport(ctx, params); setReportErr != nil {
		return nil, setReportErr
	}

	return &rezai.SaveSituationInvestigationReportToolOutput{Saved: true}, nil
}
