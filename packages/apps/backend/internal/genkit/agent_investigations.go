package genkit

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/investigation"
	siti "github.com/rezible/rezible/ent/situationinvestigation"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type InvestigationAgent struct {
	investigations rez.InvestigationService
	situations     rez.SituationService
	analyses       rez.SystemAnalysisService
	knowledge      rez.KnowledgeGraphService
}

func NewInvestigationAgent(investigations rez.InvestigationService, situations rez.SituationService, analyses rez.SystemAnalysisService, knowledge rez.KnowledgeGraphService) *InvestigationAgent {
	return &InvestigationAgent{investigations: investigations, situations: situations, analyses: analyses, knowledge: knowledge}
}

func (a *InvestigationAgent) agentDefinition() rezai.InvestigationAgentDefinition {
	return rezai.InvestigationAgent
}

func (a *InvestigationAgent) makeMiddleware() []ai.Middleware {
	resolveInvocationSystemAnalysisID := func(ctx context.Context) (uuid.UUID, error) {
		invCtx := getAgentInvocationContext(ctx)
		if invCtx == nil {
			return uuid.Nil, fmt.Errorf("agent session context does not exist")
		}
		inv, invErr := a.investigations.LookupInvestigation(ctx, investigation.AgentSessionID(invCtx.Session.ID))
		if invErr != nil {
			return uuid.Nil, invErr
		}
		return inv.SystemAnalysisID, nil
	}
	return []ai.Middleware{
		newSystemAnalysisMiddleware(a.analyses, a.knowledge, resolveInvocationSystemAnalysisID),
		newSituationInvestigationReportMiddleware(a.situations),
	}
}

func (a *InvestigationAgent) makeInitialTurnInput(ctx context.Context, input rezai.InvestigationAgentSessionInput) (*rez.AiAgentTurnInput, error) {
	return &rez.AiAgentTurnInput{}, nil
}

func (a *InvestigationAgent) updateInitialTurnMessage(ctx context.Context, input rezai.InvestigationAgentSessionInput) (string, error) {
	if input.Query != nil && *input.Query != "" {
		return *input.Query, nil
	}
	return "Investigate the available evidence and report the likely cause and best next step.", nil
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

func newSituationInvestigationReportMiddleware(situations rez.SituationService) *situationInvestigationReportMiddleware {
	return &situationInvestigationReportMiddleware{situations: situations}
}

func (m *situationInvestigationReportMiddleware) Name() string {
	return "situation_investigation_report"
}

func (m *situationInvestigationReportMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	invCtx := getAgentInvocationContext(ctx)
	if invCtx == nil {
		return nil, fmt.Errorf("agent session context does not exist")
	}

	sitInvPred := siti.HasInvestigationWith(investigation.AgentSessionID(invCtx.Session.ID))
	sitInv, lookupSitInvErr := m.situations.LookupSituationInvestigation(ctx, sitInvPred)
	if lookupSitInvErr != nil && !ent.IsNotFound(lookupSitInvErr) {
		return nil, fmt.Errorf("get investigation for agent session: %w", lookupSitInvErr)
	}
	if sitInv == nil {
		return &ai.Hooks{}, nil
	}

	isRequestedTurn := sitInv.RequestedTurnID != nil && *sitInv.RequestedTurnID == invCtx.Turn.ID
	if !isRequestedTurn {
		reportText := "No investigation report has been accepted yet."
		if report := sitInv.Edges.Investigation.Edges.Report; report != nil {
			reportText = "The existing investigation report is: " + report.Text
		}
		return &ai.Hooks{
			WrapGenerate: makeSystemTextInjectorFn("situation_report", "This is a conversational follow-up.\n"+reportText),
		}, nil
	}

	return &ai.Hooks{
		Tools:        []ai.Tool{m.makeUpdateReportTool()},
		WrapGenerate: m.makeGenerateWrapper(invCtx.Turn.ID, sitInv),
	}, nil
}

func (m *situationInvestigationReportMiddleware) makeGenerateWrapper(turnId uuid.UUID, sitInv *ent.SituationInvestigation) wrapGenerateFn {
	return func(ctx context.Context, params *ai.GenerateParams, next ai.GenerateNext) (*ai.ModelResponse, error) {
		response, respErr := next(ctx, params)
		if respErr != nil {
			return nil, respErr
		}
		//if response != nil && len(response.ToolRequests()) == 0 {
		//	current, getInvErr := m.situations.LookupSituationInvestigation(ctx, siti.SituationID(situationId))
		//	if getInvErr != nil {
		//		return nil, getInvErr
		//	}
		//	if current.RequestedTurnID == nil || *current.RequestedTurnID != turnId || current.CompletedRevision < current.RequestedRevision {
		//		return nil, fmt.Errorf("investigation finished without an accepted report")
		//	}
		//}
		return response, nil
	}
}

func (m *situationInvestigationReportMiddleware) makeUpdateReportTool() ai.Tool {
	return makeDefinedTool(rezai.SaveSituationInvestigationReportTool, func(ctx context.Context, input rezai.SaveSituationInvestigationReportToolInput) (*rezai.SaveSituationInvestigationReportToolOutput, error) {
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
	})
}
