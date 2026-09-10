package genkit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type InvestigationAgent struct {
	situations rez.SituationService
}

func NewInvestigationAgent(situations rez.SituationService) *InvestigationAgent {
	return &InvestigationAgent{situations: situations}
}

func (a *InvestigationAgent) agentDefinition() rezai.InvestigationAgentDefinition {
	return rezai.InvestigationAgent
}

func (a *InvestigationAgent) makeInitialTurnInput(ctx context.Context, input rezai.InvestigationAgentSessionInput) (*rez.AiAgentTurnInput, error) {
	return &rez.AiAgentTurnInput{
		Message: ai.NewUserTextMessage("Investigate this situation."),
	}, nil
}

func (a *InvestigationAgent) updateInitialTurnMessage(ctx context.Context, input rezai.InvestigationAgentSessionInput) (string, error) {
	sit, situationErr := a.situations.GetSituation(ctx, input.SituationID)
	if situationErr != nil {
		return "", fmt.Errorf("get situation: %w", situationErr)
	}

	return fmt.Sprintf(`Title: %s
Summary: %s
Status: %s
Opened at: %s`, sit.Title, sit.Summary, sit.Status, sit.OpenedAt.Format("2006-01-02T15:04:05Z07:00")), nil
}

func (a *InvestigationAgent) makeMiddleware() []ai.Middleware {
	return []ai.Middleware{
		&situationInvestigationReportMiddleware{situations: a.situations},
	}
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
	invCtx, ok := getAgentInvocationContext(ctx)
	if !ok || invCtx.Turn == nil {
		return nil, fmt.Errorf("missing agent invocation context")
	}
	var input rezai.InvestigationAgentSessionInput
	if jsonErr := json.Unmarshal(invCtx.Session.Input, &input); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal session input: %w", jsonErr)
	}

	inv, invErr := m.situations.GetSituationInvestigation(ctx, input.InvestigationID)
	if invErr != nil {
		return nil, fmt.Errorf("get investigation: %w", invErr)
	}
	hooks := &ai.Hooks{}

	requested := inv.RequestedTurnID != nil && *inv.RequestedTurnID == invCtx.Turn.ID
	if !requested {
		hooks.WrapGenerate = makeSystemTextInjectorFn("situation_report", "This is a conversational follow-up. The existing investigation report is: "+inv.Report.Text)
	} else {
		hooks.Tools = append(hooks.Tools, makeDefinedTool(rezai.SaveSituationInvestigationReportTool, m.updateReportToolFunc))
		hooks.WrapGenerate = m.makeGenerateWrapper(invCtx.Turn.ID, input.SituationID, input.InvestigationID)
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
			current, getInvErr := m.situations.GetSituationInvestigation(ctx, investigationID)
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
	invocation, ok := getAgentInvocationContext(ctx)
	if !ok || invocation.Turn == nil {
		return nil, fmt.Errorf("missing agent invocation context")
	}

	params := rez.SetSituationInvestigationReportParams{
		AgentTurnID: invocation.Turn.ID,
		Report:      input.Report,
		Assessments: input.Assessments,
	}
	if _, setReportErr := m.situations.SetSituationInvestigationReport(ctx, params); setReportErr != nil {
		return nil, setReportErr
	}

	return &rezai.SaveSituationInvestigationReportToolOutput{Saved: true}, nil
}
