package evals

import (
	"context"
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
	report, artifactCheck := decodeSituationInvestigationReport(result)
	if report == nil {
		return rezai.EvalScenarioGrade{Checks: []rezai.EvalCheck{artifactCheck}}, nil
	}

	state, stateErr := queryAnalysisState(ctx, client, s.fixture.analysisID)
	if stateErr != nil {
		return rezai.EvalScenarioGrade{}, stateErr
	}

	limitationsPassed := len(report.Limitations) > 0
	limitationsSummary := "The investigation report explicitly records missing context."
	if !limitationsPassed {
		limitationsSummary = "The investigation report contains no nonblank limitations."
	}
	limitationsCheck := rezai.EvalCheck{
		ID:       "limitations",
		Passed:   limitationsPassed,
		Summary:  limitationsSummary,
		Expected: "at least one nonblank limitation",
		Observed: report.Limitations,
	}

	noFindingsPassed := len(state.findings) == 0
	noFindingsSummary := "The analysis contains no unsupported durable findings."
	if !noFindingsPassed {
		noFindingsSummary = "The analysis contains unsupported durable findings."
	}
	noFindingsCheck := rezai.EvalCheck{
		ID:       "no_unsupported_findings",
		Passed:   noFindingsPassed,
		Summary:  noFindingsSummary,
		Expected: []subjectObservation{},
		Observed: state.findings,
	}

	checks := []rezai.EvalCheck{
		artifactCheck,
		reportTextCheck(report),
		limitationsCheck,
		nextActionCheck(report),
		noFindingsCheck,
	}
	return rezai.EvalScenarioGrade{Output: report, Checks: checks}, nil
}
