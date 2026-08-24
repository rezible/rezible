package evals

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	genkitai "github.com/firebase/genkit/go/ai"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/schema/schematypes"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type AlertsInsufficientContext struct{}

func (*AlertsInsufficientContext) Definition() rezai.EvalScenarioDefinition {
	return rezai.EvalScenarioDefinition{
		Name:        "alerts/insufficient-context",
		Description: "The alerts agent should acknowledge missing evidence and recommend useful next checks.",
		AgentName:   rezai.AlertsAgent.Name,
	}
}

func (*AlertsInsufficientContext) Seed(ctx context.Context, client *ent.Client) (rezai.EvalScenarioSeed, error) {
	entity, entityErr := client.KnowledgeEntity.Create().
		SetKind(kne.KindSignal).
		SetSubkind("alert").
		Save(ctx)
	if entityErr != nil {
		return rezai.EvalScenarioSeed{}, fmt.Errorf("create alert knowledge entity: %w", entityErr)
	}

	alert, alertErr := client.Alert.Create().
		SetTitle("Checkout API error rate is high").
		SetDescription("The checkout API error rate exceeded its warning threshold.").
		SetDefinition("sum(rate(checkout_http_requests_total{status=~\"5..\"}[5m])) > 1").
		SetKnowledgeEntityID(entity.ID).
		Save(ctx)
	if alertErr != nil {
		return rezai.EvalScenarioSeed{}, fmt.Errorf("create alert: %w", alertErr)
	}

	instance, instanceErr := client.AlertInstance.Create().
		SetAlertID(alert.ID).
		Save(ctx)
	if instanceErr != nil {
		return rezai.EvalScenarioSeed{}, fmt.Errorf("create alert instance: %w", instanceErr)
	}

	analysis, analysisErr := client.SystemAnalysis.Create().
		SetSubjectEntityID(entity.ID).
		Save(ctx)
	if analysisErr != nil {
		return rezai.EvalScenarioSeed{}, fmt.Errorf("create system analysis: %w", analysisErr)
	}
	analysisEntityErr := client.SystemAnalysisEntity.Create().
		SetAnalysisID(analysis.ID).
		SetKnowledgeEntityID(entity.ID).
		Exec(ctx)
	if analysisEntityErr != nil {
		return rezai.EvalScenarioSeed{}, fmt.Errorf("include analysis entity: %w", analysisEntityErr)
	}

	return rezai.EvalScenarioSeed{
		Input: rezai.AlertAgentInput{
			AlertInstanceID: instance.ID,
		},
		SystemAnalysisID: &analysis.ID,
	}, nil
}

func (s *AlertsInsufficientContext) Judge(_ context.Context, _ *ent.Client, result *rez.AiAgentInvocationResult) ([]genkitai.Score, error) {
	report, reportErr := s.investigationReport(result)
	makeScore := func(id string, passed bool, detail string) genkitai.Score {
		status := genkitai.ScoreStatusFail.String()
		if passed {
			status = genkitai.ScoreStatusPass.String()
		}
		return genkitai.Score{Id: id, Score: passed, Status: status, Details: map[string]any{"detail": detail}}
	}
	reportDetail := "investigation report artifact is valid"
	if reportErr != nil {
		reportDetail = reportErr.Error()
	}
	checks := []genkitai.Score{
		makeScore("report_artifact", reportErr == nil, reportDetail),
	}
	if reportErr != nil {
		return checks, nil
	}

	checks = append(checks,
		makeScore("report_text", strings.TrimSpace(report.Text) != "", "investigation report contains user-facing text"),
		makeScore("limitations", len(report.Limitations) > 0, "investigation report explicitly records missing context"),
		makeScore("next_action", len(report.SuggestedChecks)+len(report.RecommendedActions) > 0 || strings.TrimSpace(report.BestNextStep) != "", "investigation report proposes a next check or action"),
	)
	return checks, nil
}

func (*AlertsInsufficientContext) investigationReport(result *rez.AiAgentInvocationResult) (*schematypes.AlertInvestigationReport, error) {
	if result == nil {
		return nil, fmt.Errorf("agent returned no invocation result")
	}
	for _, artifact := range result.State.Artifacts {
		if artifact == nil || artifact.Name != "investigation_report" {
			continue
		}
		for _, part := range artifact.Parts {
			if part == nil || !strings.HasPrefix(strings.TrimSpace(part.Text), "{") {
				continue
			}
			var report schematypes.AlertInvestigationReport
			if jsonErr := json.Unmarshal([]byte(part.Text), &report); jsonErr != nil {
				return nil, fmt.Errorf("decode investigation report: %w", jsonErr)
			}
			return &report, nil
		}
		return nil, fmt.Errorf("investigation report artifact has no JSON part")
	}
	return nil, fmt.Errorf("investigation report artifact was not created")
}
