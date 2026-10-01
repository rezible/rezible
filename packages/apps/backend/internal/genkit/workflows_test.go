package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	gkai "github.com/firebase/genkit/go/ai"
	rez "github.com/rezible/rezible"
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

func (s *AiRuntimeSuite) makeWorkflowOutputModel(output *testWorkflowOutput) ModelDefinition[any] {
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
	svc := s.makeRuntime(ctx, WithDefinedModel(model))
	runner := &testWorkflowRunner{}
	builder := NewWorkflowBuilder(svc, runner)
	definition := rezai.AiPromptWorkflowDefinition[testWorkflowInput, testWorkflowOutput]{
		Name:  "test_workflow_validation",
		Model: model.Name,
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
	svc := s.makeRuntime(ctx, WithDefinedModel(model))
	runner := &testWorkflowRunner{}
	builder := NewWorkflowBuilder(svc, runner)
	definition := rezai.AiPromptWorkflowDefinition[testWorkflowInput, testWorkflowOutput]{
		Name:         "test_workflow",
		Model:        model.Name,
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

func (s *AiRuntimeSuite) TestDefineWorkflowRejectsDuplicateNames() {
	ctx := execution.NewSystemContext(s.T().Context())

	model := s.makeWorkflowOutputModel(&testWorkflowOutput{OK: true})
	svc := s.makeRuntime(ctx, WithDefinedModel(model))
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
