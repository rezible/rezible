package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	rez "github.com/rezible/rezible"
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
		Message: ai.NewUserTextMessage(fmt.Sprintf("Investigate alert instance %s.", input.AlertInstanceID)),
	}, nil
}

func (a *AlertsAgent) makeInitialContextSeed(ctx context.Context, input rezai.AlertAgentInput) (string, error) {
	inst, instErr := a.alerts.GetAlertInstance(ctx, input.AlertInstanceID)
	if instErr != nil {
		return "", fmt.Errorf("get alert instance: %w", instErr)
	}

	alrt, alrtErr := inst.Edges.AlertOrErr()
	if alrtErr != nil {
		return "", fmt.Errorf("get alert: %w", alrtErr)
	}

	seed := fmt.Sprintf(`Alert instance ID: %s
Title: %s
Description: %s
Definition: %s`, input.AlertInstanceID, alrt.Title, alrt.Description, alrt.Definition)
	if alrt.KnowledgeEntityID != nil {
		seed += fmt.Sprintf("\nKnowledge graph entity ID: %s", *alrt.KnowledgeEntityID)
	}
	return seed, nil
}

func (a *AlertsAgent) makeMiddleware() []ai.Middleware {
	return []ai.Middleware{&alertInvestigationReportMiddleware{}}
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
		Tools: []ai.Tool{m.makeUpdateReportTool()},
	}, nil
}

func (m *alertInvestigationReportMiddleware) makeUpdateReportTool() ai.Tool {
	return aix.NewTool(
		rezai.SaveAlertInvestigationReportTool.Name(),
		rezai.SaveAlertInvestigationReportTool.Description(),
		func(ctx context.Context, input rezai.SaveAlertInvestigationReportInput) (rezai.SaveAlertInvestigationReportOutput, error) {
			report := input.Report
			report.Text = strings.TrimSpace(report.Text)
			if report.Text == "" {
				return rezai.SaveAlertInvestigationReportOutput{}, fmt.Errorf("%w: report text is required", rez.ErrInvalidInput)
			}
			reportJson, jsonErr := json.Marshal(report)
			if jsonErr != nil {
				return rezai.SaveAlertInvestigationReportOutput{}, jsonErr
			}
			as := aix.ArtifactStoreFromContext(ctx)
			as.AddArtifacts(&aix.Artifact{
				Name:  "investigation_report",
				Parts: []*ai.Part{ai.NewJSONPart(string(reportJson))},
			})
			return rezai.SaveAlertInvestigationReportOutput{Saved: true}, nil
		},
	)
}
