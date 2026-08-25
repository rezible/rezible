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

type (
	AlertsAgent struct {
		alerts rez.AlertService
	}
)

func NewAlertsAgent(alerts rez.AlertService) *AlertsAgent {
	return &AlertsAgent{alerts: alerts}
}

func (a *AlertsAgent) agentDefinition() rezai.AlertsAgentDefinition {
	return rezai.AlertsAgent
}

func (a *AlertsAgent) makeInitialTurnInput(ctx context.Context, input rezai.AlertAgentInput) (*rez.AiAgentTurnInput, error) {
	return &rez.AiAgentTurnInput{
		Message: ai.NewUserTextMessage(fmt.Sprintf("Investigate this alert instance.")),
	}, nil
}

func (a *AlertsAgent) updateInitialTurnMessage(ctx context.Context, input rezai.AlertAgentInput) (string, error) {
	inst, instErr := a.alerts.GetAlertInstance(ctx, input.AlertInstanceID)
	if instErr != nil {
		return "", fmt.Errorf("get alert instance: %w", instErr)
	}

	alrt, alrtErr := inst.Edges.AlertOrErr()
	if alrtErr != nil {
		return "", fmt.Errorf("get alert: %w", alrtErr)
	}

	seed := fmt.Sprintf(`Title: %s
Description: %s
Definition: %s`, alrt.Title, alrt.Description, alrt.Definition)
	return seed, nil
}

func (a *AlertsAgent) makeMiddleware() []ai.Middleware {
	return []ai.Middleware{
		&alertInvestigationReportMiddleware{},
	}
}

func (a *AlertsAgent) getCustomState(context.Context, *ent.AgentSession) (*rezai.AlertAgentState, error) {
	return &rezai.AlertAgentState{}, nil
}

func (a *AlertsAgent) transformState(ctx context.Context, state *aix.SessionState[rezai.AlertAgentState]) (*aix.SessionState[rezai.AlertAgentState], error) {
	return state, nil
}

func (a *AlertsAgent) transformStreamChunk(ctx context.Context, chunk *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error) {
	return chunk, nil
}

type alertInvestigationReportMiddleware struct{}

func (m *alertInvestigationReportMiddleware) Name() string {
	return "alert_investigation_report"
}

func (m *alertInvestigationReportMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		Tools: []ai.Tool{makeDefinedTool(rezai.SaveAlertInvestigationReportTool, m.updateReportToolFunc)},
	}, nil
}

func (m *alertInvestigationReportMiddleware) updateReportToolFunc(ctx context.Context, input rezai.SaveAlertInvestigationReportToolInput) (*rezai.SaveAlertInvestigationReportToolOutput, error) {
	report := input.Report
	report.Text = strings.TrimSpace(report.Text)
	if report.Text == "" {
		return nil, fmt.Errorf("%w: report text is required", rez.ErrInvalidInput)
	}
	reportJson, jsonErr := json.Marshal(report)
	if jsonErr != nil {
		return nil, jsonErr
	}
	as := aix.ArtifactStoreFromContext(ctx)
	as.AddArtifacts(&aix.Artifact{
		Name:  "investigation_report",
		Parts: []*ai.Part{ai.NewJSONPart(string(reportJson))},
	})
	return &rezai.SaveAlertInvestigationReportToolOutput{Saved: true}, nil
}
