package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	gkai "github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/schema/schematypes"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
)

type (
	testWorkflowInput struct {
		Message string `json:"message"`
	}

	testWorkflowOutput struct {
		OK bool `json:"ok"`
	}

	testWorkflowRunner struct {
		calls int
	}
)

func (i testWorkflowInput) Validate() error {
	if strings.TrimSpace(i.Message) == "" {
		return fmt.Errorf("message is required")
	}
	return nil
}

func (r *testWorkflowRunner) ExecuteWorkflow(ctx context.Context, name string, run func(context.Context) error) error {
	r.calls++
	return run(ctx)
}

func (s *AiRuntimeSuite) makeWorkflowOutputModel(output *testWorkflowOutput) testModel {
	out, jsonErr := json.Marshal(output)
	s.Require().NoError(jsonErr)
	response := &gkai.ModelResponse{
		Message: gkai.NewModelTextMessage(string(out)),
	}
	return makeTestOutputModel(response)
}

func (s *AiRuntimeSuite) TestDefinePromptWorkflowValidatesInput() {
	ctx := execution.NewSystemContext(s.T().Context())

	model := s.makeWorkflowOutputModel(&testWorkflowOutput{OK: true})
	svc := s.makeRuntime(ctx, model)
	runner := &testWorkflowRunner{}
	builder := NewWorkflowBuilder(svc, runner)
	definition := rezai.AiPromptWorkflowDefinition[testWorkflowInput, testWorkflowOutput]{
		Name: "test_workflow_validation",
		Prompt: func(input testWorkflowInput) string {
			return input.Message
		},
	}

	workflow, workflowErr := builder.DefinePromptWorkflow(definition)
	s.Require().NoError(workflowErr)

	_, runErr := workflow.Run(ctx, testWorkflowInput{})
	s.Require().ErrorContains(runErr, "message is required")
	s.Equal(1, runner.calls)
}

func (s *AiRuntimeSuite) TestDefinePromptWorkflowRunsTypedOutput() {
	ctx := execution.NewSystemContext(s.T().Context())

	model := s.makeWorkflowOutputModel(&testWorkflowOutput{OK: true})
	svc := s.makeRuntime(ctx, model)
	runner := &testWorkflowRunner{}
	builder := NewWorkflowBuilder(svc, runner)
	definition := rezai.AiPromptWorkflowDefinition[testWorkflowInput, testWorkflowOutput]{
		Name:         "test_workflow",
		SystemPrompt: "Return JSON.",
		Prompt: func(input testWorkflowInput) string {
			return "Message: " + input.Message
		},
	}

	workflow, workflowErr := builder.DefinePromptWorkflow(definition)
	s.Require().NoError(workflowErr)

	output, runErr := workflow.Run(ctx, testWorkflowInput{Message: "hello"})
	s.Require().NoError(runErr)
	s.True(output.OK, "expected output to be ok")
	s.Equal(1, runner.calls)
}

func (s *AiRuntimeSuite) TestDefinePromptWorkflowCountsUsageWithTheDefaultModel() {
	ctx := execution.NewSystemContext(s.T().Context())

	out, jsonErr := json.Marshal(testWorkflowOutput{OK: true})
	s.Require().NoError(jsonErr)
	model := makeTestOutputModel(&gkai.ModelResponse{
		Message: gkai.NewModelTextMessage(string(out)),
		Usage:   &gkai.GenerationUsage{InputTokens: 120, OutputTokens: 8, ThoughtsTokens: 4},
	})
	svc := s.makeRuntime(ctx, model)
	builder := NewWorkflowBuilder(svc, &testWorkflowRunner{})
	// No model named: the workflow runs on the default.
	definition := rezai.AiPromptWorkflowDefinition[testWorkflowInput, testWorkflowOutput]{
		Name: "test_workflow_usage",
		Prompt: func(input testWorkflowInput) string {
			return input.Message
		},
	}
	workflow, workflowErr := builder.DefinePromptWorkflow(definition)
	s.Require().NoError(workflowErr)

	tallyCtx, totals := rezai.WithTurnUsage(ctx)
	_, runErr := workflow.Run(tallyCtx, testWorkflowInput{Message: "hello"})
	s.Require().NoError(runErr)

	s.Equal(rezai.TurnUsage{Model: model.Name, InputTokens: 120, OutputTokens: 12, ModelCalls: 1, PeakInputTokens: 120}, totals())
}

func (s *AiRuntimeSuite) TestNestedTurnUsageCountsInTheInnerTallyOnly() {
	ctx := execution.NewSystemContext(s.T().Context())

	out, jsonErr := json.Marshal(testWorkflowOutput{OK: true})
	s.Require().NoError(jsonErr)
	model := makeTestOutputModel(&gkai.ModelResponse{
		Message: gkai.NewModelTextMessage(string(out)),
		Usage:   &gkai.GenerationUsage{InputTokens: 120, OutputTokens: 8},
	})
	svc := s.makeRuntime(ctx, model)
	definition := rezai.AiPromptWorkflowDefinition[testWorkflowInput, testWorkflowOutput]{
		Name: "test_workflow_nested_usage",
		Prompt: func(input testWorkflowInput) string {
			return input.Message
		},
	}
	workflow, workflowErr := NewWorkflowBuilder(svc, &testWorkflowRunner{}).DefinePromptWorkflow(definition)
	s.Require().NoError(workflowErr)

	outerCtx, outerTotals := rezai.WithTurnUsage(ctx)
	innerCtx, innerTotals := rezai.WithTurnUsage(outerCtx)
	_, runErr := workflow.Run(innerCtx, testWorkflowInput{Message: "hello"})
	s.Require().NoError(runErr)

	s.Equal(rezai.TurnUsage{Model: model.Name, InputTokens: 120, OutputTokens: 8, ModelCalls: 1, PeakInputTokens: 120}, innerTotals())
	s.Equal(rezai.TurnUsage{}, outerTotals())
}

func (s *AiRuntimeSuite) TestDefinePromptWorkflowRunsOnItsModelSetting() {
	ctx := execution.NewSystemContext(s.T().Context())

	out, jsonErr := json.Marshal(testWorkflowOutput{OK: true})
	s.Require().NoError(jsonErr)
	override := makeTestOutputModel(&gkai.ModelResponse{
		Message: gkai.NewModelTextMessage(string(out)),
		Usage:   &gkai.GenerationUsage{InputTokens: 10, OutputTokens: 2},
	})
	override.Name = "test/override"
	cfg := s.Config()
	cfg.AI.Models.Default = testModelName
	cfg.AI.Models.ThreadClassifier = override.Name
	svc := newTestRuntime(cfg, makeTestOutputModel(nil), override)
	// The default model has no answer, so the workflow only succeeds on its setting's model.
	s.Require().NoError(svc.Init(ctx))
	builder := NewWorkflowBuilder(svc, &testWorkflowRunner{})
	definition := rezai.AiPromptWorkflowDefinition[testWorkflowInput, testWorkflowOutput]{
		Name: rezai.ClassifyAgentThreadResponseDefinition.Name,
		Prompt: func(input testWorkflowInput) string {
			return input.Message
		},
	}
	workflow, workflowErr := builder.DefinePromptWorkflow(definition)
	s.Require().NoError(workflowErr)

	tallyCtx, totals := rezai.WithTurnUsage(ctx)
	output, runErr := workflow.Run(tallyCtx, testWorkflowInput{Message: "hello"})
	s.Require().NoError(runErr)
	s.True(output.OK)
	s.Equal(rezai.TurnUsage{Model: override.Name, InputTokens: 10, OutputTokens: 2, ModelCalls: 1, PeakInputTokens: 10}, totals())
}

func (s *AiRuntimeSuite) TestDefinePromptWorkflowFailsWithoutAModel() {
	ctx := execution.NewSystemContext(s.T().Context())

	svc := newTestRuntime(s.Config())
	s.Require().NoError(svc.Init(ctx), "a runtime without a model starts")
	workflow, workflowErr := NewWorkflowBuilder(svc, &testWorkflowRunner{}).DefinePromptWorkflow(rezai.ClassifyAgentThreadResponseDefinition)
	s.Require().NoError(workflowErr, "a workflow without a model is defined")

	_, runErr := workflow.Run(ctx, rezai.ClassifyAgentThreadResponseInput{PreviousMessages: []string{}, UserMessage: "hello"})
	s.ErrorIs(runErr, rezai.ErrNoModel)
}

func (s *AiRuntimeSuite) TestDefineWorkflowRejectsDuplicateNames() {
	ctx := execution.NewSystemContext(s.T().Context())

	model := s.makeWorkflowOutputModel(&testWorkflowOutput{OK: true})
	svc := s.makeRuntime(ctx, model)
	builder := NewWorkflowBuilder(svc, &testWorkflowRunner{})
	run := func(context.Context, testWorkflowInput) (testWorkflowOutput, error) {
		return testWorkflowOutput{}, nil
	}

	_, firstErr := builder.DefineWorkflow("duplicate_workflow", run)
	_, secondErr := builder.DefineWorkflow("duplicate_workflow", run)
	s.Require().NoError(firstErr)
	s.Require().ErrorContains(secondErr, "already defined")
}

var _ rez.AiWorkflow[testWorkflowInput, testWorkflowOutput] = (*typedWorkflow[testWorkflowInput, testWorkflowOutput])(nil)

func (s *AiRuntimeSuite) TestJudgeSituationCandidateWorkflowAcceptsFactsAndRejectsMalformedAnswers() {
	ctx := execution.NewSystemContext(s.T().Context())
	entityID := uuid.New()
	input := rezai.SituationJudgeInput{
		Facts: schematypes.SituationFacts{
			Signals: []schematypes.SituationSignalFacts{{
				Ref:            "s1",
				EntityID:       uuid.New(),
				SourceEntityID: new(uuid.New()),
				EntityIDs:      []uuid.UUID{entityID},
				Alert: &schematypes.SituationAlertFacts{
					Severity:        schematypes.SignalSeverityCritical,
					ActiveInstances: []schematypes.SituationAlertInstanceFacts{},
				},
			}},
			LinkedIncidents: []schematypes.SituationIncidentFacts{},
			Entities:        []schematypes.SituationEntityFacts{{ID: entityID, Properties: map[string]any{"tier": "core"}, Matching: true}},
			Relationships:   []schematypes.SituationRelationshipFacts{},
			Earlier:         []schematypes.SituationEarlierFacts{},
		},
		Reasons:   []schematypes.SituationReasonResult{{Reason: schematypes.SituationRaiseReasonBreadth, Met: true}},
		Truncated: []string{},
	}
	answers := []struct {
		name      string
		response  string
		malformed bool
	}{
		{name: "decision", response: `{"decision": "hold", "reasons": [], "explanation": "Not enough evidence."}`},
		{name: "malformed", response: `{"decision": "raise", "reasons": "breadth"}`, malformed: true},
	}
	for _, answer := range answers {
		s.Run(answer.name, func() {
			response := &gkai.ModelResponse{Message: gkai.NewModelTextMessage(answer.response)}
			svc := s.makeRuntime(ctx, makeTestOutputModel(response))
			builder := NewWorkflowBuilder(svc, &testWorkflowRunner{})
			workflow, workflowErr := builder.DefinePromptWorkflow(rezai.JudgeSituationCandidateDefinition)
			s.Require().NoError(workflowErr)

			output, runErr := workflow.Run(ctx, input)
			if answer.malformed {
				s.Require().ErrorIs(runErr, rezai.ErrWorkflowInvalidOutput)
				return
			}
			s.Require().NoError(runErr, "the facts, with their UUIDs and empty lists, match the workflow's input schema")
			s.Equal("hold", output.Decision)
		})
	}
}
