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
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/firebase/genkit/go/ai"

	"github.com/rezible/rezible/internal/db/evals"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/integrations"
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
	// sources are what the first turn was told it can read.
	sources *rezai.InvestigationSources
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
	if m.sources == nil {
		m.sources = &introduction.Sources
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

// runScenario runs one repetition of the scenario in a new application, with the model action playing the
// agent.
func (s *BackendSuite) runScenario(name string, generate ai.ModelActionFunc[any]) evals.Repetition {
	ah := s.newAppHarness(appTestOptions{ModelAction: generate, AgentEvaluation: true})
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
		model := &scenarioModel{cited: mapset.NewSet[string]()}
		repetition := s.runScenario("alerts-without-sources", model.generate)
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
		s.Equal(&rezai.InvestigationSources{
			Logs:    rezai.SourceAvailability{Reason: "no logs integration installed"},
			Changes: rezai.SourceAvailability{Available: true},
		}, model.sources)
	})

	s.Run("a failed turn", func() {
		model := &scenarioModel{cited: mapset.NewSet[string](), loop: true}
		repetition := s.runScenario("alerts-without-sources", model.generate)
		s.Equal(evals.RunStatusError, repetition.Status)
		s.Require().NotNil(repetition.Error)
		s.Equal("setup", repetition.Error.Step)
		s.Contains(repetition.Error.Message, "turn 1 failed")
	})
}

// checkoutModel plays the investigation agent in checkout-upstream-timeout. Its first turn finds both
// services by name, lists checkout-api's changes and searches both services' logs, then publishes a report and a
// finding citing the deployment and the checkout-api read, and a hypothesis citing the payments-api read. It
// answers the question citing the payments-api read.
type checkoutModel struct {
	investigationModel
	checkoutSearch *rezai.SearchLogsToolResult
	paymentsSearch *rezai.SearchLogsToolResult
}

var (
	foundServicePattern  = regexp.MustCompile(`^entity (\S+) `)
	deploymentEvidence   = regexp.MustCompile(`deployment .*version=checkout-api@2041 .*latest_evidence=(\S+)`)
	checkoutServiceNames = []string{"checkout-api", "payments-api"}
)

func (m *checkoutModel) generate(_ context.Context, request *ai.ModelRequest, _ any, _ ai.ModelStreamCallback) (*ai.ModelResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.step++
	stop := &ai.ModelResponse{Message: ai.NewModelTextMessage("Recorded."), FinishReason: ai.FinishReasonStop}
	introduction, introductionErr := m.decodeIntroduction(request)
	if introductionErr != nil {
		return nil, introductionErr
	}
	call := func(name, ref string, input any) *ai.Part {
		return ai.NewToolRequestPart(&ai.ToolRequest{Name: name, Ref: ref, Input: input})
	}

	last := request.Messages[len(request.Messages)-1]
	if last.Role != ai.RoleTool {
		if introduction.HasAssignedUserQuestion {
			answer := rezai.PublishInvestigationAnswerToolInput{
				Title: "Yes", Body: "payments-api is exhausting its connection pool.",
				EvidenceRefs: []string{m.paymentsSearch.ProviderReadRef},
			}
			return &ai.ModelResponse{Message: ai.NewModelMessage(call(rezai.PublishInvestigationAnswerTool.Name(), "answer", answer))}, nil
		}
		parts := make([]*ai.Part, len(checkoutServiceNames))
		for index, name := range checkoutServiceNames {
			parts[index] = call(rezai.ListServicesTool.Name(), "find-"+name, rezai.ListServicesArgs{Name: name})
		}
		return &ai.ModelResponse{Message: ai.NewModelMessage(parts...)}, nil
	}

	switch last.Content[0].ToolResponse.Name {
	case rezai.ListServicesTool.Name():
		refs := make(map[string]string, len(checkoutServiceNames))
		for _, name := range checkoutServiceNames {
			var found rezai.TextToolResult
			if decodeErr := m.decodeToolResult(request, rezai.ListServicesTool.Name(), "find-"+name, &found); decodeErr != nil {
				return nil, decodeErr
			}
			match := foundServicePattern.FindStringSubmatch(found.Text)
			if match == nil {
				return nil, fmt.Errorf("step %d: no service in %q", m.step, found.Text)
			}
			refs[name] = match[1]
		}
		start := introduction.TurnStartedAt.Add(-time.Hour).Format(time.RFC3339)
		end := introduction.TurnStartedAt.Format(time.RFC3339)
		changes := rezai.ListChangesArgs{ServiceRef: refs["checkout-api"], Start: start, End: end}
		checkoutLogs := rezai.SearchLogsArgs{ServiceRef: refs["checkout-api"], Start: start, End: end, Contains: "upstream timeout"}
		paymentsLogs := rezai.SearchLogsArgs{ServiceRef: refs["payments-api"], Start: start, End: end, Contains: "pool exhausted"}
		return &ai.ModelResponse{Message: ai.NewModelMessage(
			call(rezai.ListChangesTool.Name(), "changes", changes),
			call(rezai.SearchLogsTool.Name(), "logs-checkout", checkoutLogs),
			call(rezai.SearchLogsTool.Name(), "logs-payments", paymentsLogs),
		)}, nil
	case rezai.ListChangesTool.Name():
		var changes rezai.TextToolResult
		if decodeErr := m.decodeToolResult(request, rezai.ListChangesTool.Name(), "changes", &changes); decodeErr != nil {
			return nil, decodeErr
		}
		match := deploymentEvidence.FindStringSubmatch(changes.Text)
		if match == nil {
			return nil, fmt.Errorf("step %d: no deployment in %q", m.step, changes.Text)
		}
		m.checkoutSearch, m.paymentsSearch = &rezai.SearchLogsToolResult{}, &rezai.SearchLogsToolResult{}
		if decodeErr := m.decodeToolResult(request, rezai.SearchLogsTool.Name(), "logs-checkout", m.checkoutSearch); decodeErr != nil {
			return nil, decodeErr
		}
		if decodeErr := m.decodeToolResult(request, rezai.SearchLogsTool.Name(), "logs-payments", m.paymentsSearch); decodeErr != nil {
			return nil, decodeErr
		}
		checkoutRefs := []string{match[1], m.checkoutSearch.ProviderReadRef}
		report := rezai.PublishInvestigationReportToolInput{Text: "Checkout times out on payments-api.", Summary: "Checkout times out.", EvidenceRefs: checkoutRefs}
		finding := rezai.PublishInvestigationFindingToolInput{Key: "timeouts", Title: "Upstream timeouts", Body: "Checkout times out on payments-api since the deployment.", EvidenceRefs: checkoutRefs}
		hypothesis := rezai.PublishInvestigationHypothesisToolInput{
			Key: "pool", Title: "payments-api pool exhausted", Justification: "Its logs say so.", Status: "supported",
			EvidenceRefs: []string{m.paymentsSearch.ProviderReadRef},
		}
		return &ai.ModelResponse{Message: ai.NewModelMessage(
			call(rezai.PublishInvestigationReportTool.Name(), "report", report),
			call(rezai.PublishInvestigationFindingTool.Name(), "finding", finding),
			call(rezai.PublishInvestigationHypothesisTool.Name(), "hypothesis", hypothesis),
		)}, nil
	default:
		return stop, nil
	}
}

func (s *BackendSuite) TestAiEvalScenarioReads() {
	model := &checkoutModel{}
	repetition := s.runScenario("checkout-upstream-timeout", model.generate)
	s.Require().Nil(repetition.Error)
	for _, check := range repetition.Checks {
		s.True(check.Passed, "%s: %s", check.ID, check.Summary)
	}
	s.Len(repetition.Checks, 9)
	s.Require().Len(repetition.Turns, 2)
	s.Equal(2, repetition.Turns[0].Reads)

	s.Require().NotNil(model.checkoutSearch)
	s.NotEmpty(model.checkoutSearch.Lines)
	s.Equal("service:checkout-api", model.checkoutSearch.ServiceName)
	s.Equal(map[string]string{"level": "ERROR"}, model.checkoutSearch.Labels)
	s.Require().Len(model.checkoutSearch.Lines, 1, "the repeated line collapses into one entry")
	s.Equal(integrations.DefaultLogSearchLimit, model.checkoutSearch.Lines[0].Count)
	s.NotEmpty(model.checkoutSearch.Limits, "the evaluation reader records the cut")
	s.NotEmpty(model.paymentsSearch.Lines)
}
