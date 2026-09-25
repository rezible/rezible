package evals

import (
	"context"
	"strings"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type InvestigationInsufficientContext struct {
	fixture investigationFixture
}

func (*InvestigationInsufficientContext) Definition() rezai.EvalScenarioDefinition {
	return rezai.EvalScenarioDefinition{
		Name:        "investigation/insufficient-context",
		AgentName:   rezai.InvestigationAgent.Name,
		Description: "The investigation agent should acknowledge missing evidence and recommend useful next checks without recording unsupported findings.",
	}
}

func (s *InvestigationInsufficientContext) Seed(ctx context.Context, client *ent.Client) (rezai.EvalScenarioSeed, error) {
	referenceTime := time.Now().UTC().Truncate(time.Second)
	fixture, seed, seedErr := seedBaseInvestigation(ctx, client, referenceTime)
	if seedErr != nil {
		return rezai.EvalScenarioSeed{}, seedErr
	}
	s.fixture = fixture
	return seed, nil
}

func (s *InvestigationInsufficientContext) Grade(ctx context.Context, client *ent.Client, result *rez.AiAgentInvocationResult) (rezai.EvalScenarioGrade, error) {
	report, reportCheck, err := loadSituationInvestigationReport(ctx, client, s.fixture.situationID)
	if err != nil {
		return rezai.EvalScenarioGrade{}, err
	}
	if report == nil {
		return rezai.EvalScenarioGrade{Checks: []rezai.EvalCheck{reportCheck}}, nil
	}

	reportText := strings.ToLower(report.Text)
	limitationsPassed := strings.Contains(reportText, "insufficient") || strings.Contains(reportText, "not available") || strings.Contains(reportText, "no evidence") || strings.Contains(reportText, "cannot determine") || strings.Contains(reportText, "not enough")
	limitationsSummary := "The investigation report explicitly records missing context."
	if !limitationsPassed {
		limitationsSummary = "The investigation report does not explain the missing context."
	}
	limitationsCheck := rezai.EvalCheck{
		ID:       "limitations",
		Passed:   limitationsPassed,
		Summary:  limitationsSummary,
		Expected: "a plain-text explanation that context is insufficient",
		Observed: report.Text,
	}

	checks := []rezai.EvalCheck{
		reportCheck,
		reportTextCheck(report),
		limitationsCheck,
	}
	return rezai.EvalScenarioGrade{Output: report, Checks: checks}, nil
}
