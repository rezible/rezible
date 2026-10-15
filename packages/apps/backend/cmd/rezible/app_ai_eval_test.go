package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/firebase/genkit/go/ai"

	"github.com/rezible/rezible/internal/db/evals"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
)

// TestAiEvalWorkflow runs the situation judge's dataset through the application, where the judge is disabled
// by configuration, and reads the results the run writes. Grading is tested in internal/db/evals; the model gives
// the same valid answer to every case.
func (s *BackendSuite) TestAiEvalWorkflow() {
	ah := s.newAppHarness(appTestOptions{
		ModelAction: func(context.Context, *ai.ModelRequest, any, ai.ModelStreamCallback) (*ai.ModelResponse, error) {
			return &ai.ModelResponse{
				Message: ai.NewModelTextMessage(`{"decision": "hold", "reasons": [], "explanation": "Weak evidence."}`),
				Usage:   &ai.GenerationUsage{InputTokens: 100, OutputTokens: 10},
			}, nil
		},
	})

	targets, targetsErr := ah.app.invoke[[]evals.WorkflowTarget]()
	s.Require().NoError(targetsErr)
	index := slices.IndexFunc(targets, func(target evals.WorkflowTarget) bool {
		return target.Target().Name == rezai.JudgeSituationCandidateDefinition.Name
	})
	s.Require().GreaterOrEqual(index, 0, "the judge is a target while disabled")
	outDir := s.T().TempDir()
	_, reportErr := evals.Report(evals.RunWorkflow(s.T().Context(), targets[index], 2), outDir, io.Discard)
	s.Require().NoError(reportErr)

	var first evals.RunResult
	s.readJSON(filepath.Join(outDir, "repetition-1.json"), &first)
	s.Require().Len(first.Repetitions, 1)
	cases := len(first.Repetitions[0].Checks)
	s.NotZero(cases, "every case is recorded as a check")

	var summary evals.Summary
	s.readJSON(filepath.Join(outDir, "summary.json"), &summary)
	s.Equal(2, summary.Runs)
	s.Zero(summary.Errors)
	s.Len(summary.Checks, cases)
	for _, repetition := range summary.Repetitions {
		s.Equal([]evals.ModelUsage{
			{Model: "test/application", InputTokens: 100 * cases, OutputTokens: 10 * cases},
		}, repetition.Usage, "the workflow's tokens are counted with its model")
	}
}

func (s *BackendSuite) readJSON(path string, into any) {
	encoded, readErr := os.ReadFile(path)
	s.Require().NoError(readErr)
	s.Require().NoError(json.Unmarshal(encoded, into))
}

// scenarioModel plays the investigation agent in an evaluation scenario. Each turn it cites the evidence of a
// signal it has not cited yet, or answers the question it was assigned.
type scenarioModel struct {
	// investigationModel supplies the step counter and its request helpers.
	investigationModel
	cited mapset.Set[string]
	// loop calls a tool on every step, so the turn reaches its tool limit and fails.
	loop bool
}

var latestEvidencePattern = regexp.MustCompile(`assertion=alert_definition_observed .*latest_evidence=(\S+)`)

func (m *scenarioModel) generate(_ context.Context, request *ai.ModelRequest, _ any, _ ai.ModelStreamCallback) (*ai.ModelResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.step++
	if m.loop {
		call := &ai.ToolRequest{Name: rezai.ReadEntityTool.Name(), Ref: fmt.Sprintf("read-%d", m.step), Input: rezai.ReadEntityArgs{Ref: "unknown"}}
		return m.requestTool(request, call)
	}
	stop := &ai.ModelResponse{Message: ai.NewModelTextMessage("Recorded."), FinishReason: ai.FinishReasonStop}

	if last := request.Messages[len(request.Messages)-1]; last.Role == ai.RoleTool {
		var evidence rezai.TextToolResult
		if m.decodeToolResult(request, rezai.ListEntityEvidenceTool.Name(), fmt.Sprintf("evidence-%d", m.step-1), &evidence) != nil {
			return stop, nil
		}
		match := latestEvidencePattern.FindStringSubmatch(evidence.Text)
		if match == nil {
			return nil, fmt.Errorf("step %d: no alert evidence in %q", m.step, evidence.Text)
		}
		refs := []string{match[1]}
		report := rezai.PublishInvestigationReportToolInput{Text: "Checkout is failing.", Summary: "Checkout is failing.", EvidenceRefs: refs}
		finding := rezai.PublishInvestigationFindingToolInput{Key: fmt.Sprintf("alert-%d", m.step), Title: "An alert fired", Body: "An alert on checkout-api fired.", EvidenceRefs: refs}
		return &ai.ModelResponse{Message: ai.NewModelMessage(
			ai.NewToolRequestPart(&ai.ToolRequest{Name: rezai.PublishInvestigationReportTool.Name(), Ref: "report", Input: report}),
			ai.NewToolRequestPart(&ai.ToolRequest{Name: rezai.PublishInvestigationFindingTool.Name(), Ref: "finding", Input: finding}),
		)}, nil
	}

	introduction, introductionErr := m.decodeIntroduction(request)
	if introductionErr != nil {
		return nil, introductionErr
	}
	if introduction.HasAssignedUserQuestion {
		answer := rezai.PublishInvestigationAnswerToolInput{Title: "No", Body: "Nothing points at search-api.", EvidenceRefs: []string{}}
		return m.requestTool(request, &ai.ToolRequest{Name: rezai.PublishInvestigationAnswerTool.Name(), Ref: "answer", Input: answer})
	}
	for _, signal := range introduction.Situation.Signals {
		if m.cited.Add(signal.Ref) {
			args := rezai.ListEntityEvidenceArgs{Ref: signal.Ref}
			call := &ai.ToolRequest{Name: rezai.ListEntityEvidenceTool.Name(), Ref: fmt.Sprintf("evidence-%d", m.step), Input: args}
			return m.requestTool(request, call)
		}
	}
	return stop, nil
}

// runScenario runs one repetition of the scenario in a new application.
func (s *BackendSuite) runScenario(name string, model *scenarioModel) evals.Repetition {
	model.cited = mapset.NewSet[string]()
	ah := s.newAppHarness(appTestOptions{ModelAction: model.generate, AgentEvaluation: true})
	agents, agentsErr := ah.app.invoke[*evals.AgentEvalService]()
	s.Require().NoError(agentsErr)
	ctx := execution.NewRootContext(s.T().Context(), execution.KindAnonymous, execution.SourceCLI)
	result, runErr := agents.Run(ctx, name, 1)
	s.Require().NoError(runErr)
	s.Require().Len(result.Repetitions, 1)
	return result.Repetitions[0]
}

func (s *BackendSuite) TestAiEvalScenario() {
	s.Run("setup, new evidence and a question", func() {
		repetition := s.runScenario("alerts-without-sources", &scenarioModel{})
		s.Require().Nil(repetition.Error)
		outcomes := make(map[string]bool, len(repetition.Checks))
		for _, check := range repetition.Checks {
			outcomes[check.ID] = check.Passed
		}
		s.Equal(map[string]bool{
			"turn_completed:1": true,
			"report_published": true,
			"outputs_cite":     true,
			"cites_entity:alert:CheckoutHighErrorRate": true,
			"turn_completed:2":                         true,
			"outputs_cite@phase2":                      true,
			"cites_entity:alert:CheckoutLatencyHigh":   true,
			"turn_completed:3":                         true,
			"answer_published":                         true,
		}, outcomes)
		s.Len(repetition.Turns, 3)
	})

	s.Run("a failed turn", func() {
		repetition := s.runScenario("alerts-without-sources", &scenarioModel{loop: true})
		s.Equal(evals.RunStatusError, repetition.Status)
		s.Require().NotNil(repetition.Error)
		s.Equal("setup", repetition.Error.Step)
		s.Contains(repetition.Error.Message, "turn 1 failed")
	})
}
