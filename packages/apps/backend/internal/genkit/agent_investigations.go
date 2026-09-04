package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
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

func (a *InvestigationAgent) makeInitialTurnInput(ctx context.Context, input rezai.InvestigationAgentInput) (*rez.AiAgentTurnInput, error) {
	return &rez.AiAgentTurnInput{
		Message: ai.NewUserTextMessage("Investigate this situation."),
	}, nil
}

func (a *InvestigationAgent) updateInitialTurnMessage(ctx context.Context, input rezai.InvestigationAgentInput) (string, error) {
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
		&situationInvestigationReportMiddleware{},
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

type situationInvestigationReportMiddleware struct{}

func (m *situationInvestigationReportMiddleware) Name() string {
	return "situation_investigation_report"
}

func (m *situationInvestigationReportMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		Tools: []ai.Tool{
			makeDefinedTool(rezai.SaveSituationInvestigationReportTool, m.updateReportToolFunc),
		},
	}, nil
}

func (m *situationInvestigationReportMiddleware) updateReportToolFunc(ctx context.Context, input rezai.SaveSituationInvestigationReportToolInput) (*rezai.SaveSituationInvestigationReportToolOutput, error) {
	report := input.Report
	report.Text = strings.TrimSpace(report.Text)
	if report.Text == "" {
		return nil, fmt.Errorf("%w: report text is required", rez.ErrInvalidInput)
	}
	reportJSON, jsonErr := json.Marshal(report)
	if jsonErr != nil {
		return nil, jsonErr
	}
	as := aix.ArtifactStoreFromContext(ctx)
	as.AddArtifacts(&aix.Artifact{
		Name:  "situation_investigation_report",
		Parts: []*ai.Part{ai.NewJSONPart(string(reportJSON))},
	})
	return &rezai.SaveSituationInvestigationReportToolOutput{Saved: true}, nil
}
