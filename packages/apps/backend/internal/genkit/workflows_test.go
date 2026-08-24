package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	gkai "github.com/firebase/genkit/go/ai"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type (
	testWorkflowInput struct {
		Message string `json:"message"`
	}

	testWorkflowOutput struct {
		OK bool `json:"ok"`
	}
)

func (i testWorkflowInput) Validate() error {
	if strings.TrimSpace(i.Message) == "" {
		return fmt.Errorf("message is required")
	}
	return nil
}

func (s *AiServiceSuite) newTestOutputModel(output *testWorkflowOutput) ModelDefinition[any] {
	out, jsonErr := json.Marshal(output)
	s.Require().NoError(jsonErr)
	response := &gkai.ModelResponse{
		Message: gkai.NewModelTextMessage(string(out)),
	}
	return ModelDefinition[any]{
		Name: "test/output",
		opts: &gkai.ModelOptions{
			Supports: &gkai.ModelSupports{
				Constrained: gkai.ConstrainedSupportAll,
				Multiturn:   true,
				SystemRole:  true,
			},
		},
		fn: func(ctx context.Context, req *gkai.ModelRequest, cfg any, cb gkai.ModelStreamCallback) (*gkai.ModelResponse, error) {
			return response, nil
		},
	}
}

func (s *AiServiceSuite) TestDefineWorkflowValidatesInput() {
	ctx := s.SeedTenantContext()

	model := s.newTestOutputModel(&testWorkflowOutput{OK: true})
	workflow := rezai.WorkflowDefinition[testWorkflowInput, testWorkflowOutput]{
		Name:  "test_workflow_validation",
		Model: model.Name,
		Prompt: func(input testWorkflowInput) string {
			return input.Message
		},
	}

	svc := s.makeService(ctx, WithDefinedModel(model), WithWorkflow(workflow))

	runner, runnerErr := workflow.GetRunner(svc)
	s.Require().NoError(runnerErr)

	_, runErr := runner.Run(ctx, testWorkflowInput{})
	s.Require().ErrorContains(runErr, "message is required")
}

func (s *AiServiceSuite) TestDefineWorkflowRunsTypedOutput() {
	ctx := s.SeedTenantContext()

	model := s.newTestOutputModel(&testWorkflowOutput{OK: true})
	workflow := rezai.WorkflowDefinition[testWorkflowInput, testWorkflowOutput]{
		Name:         "test_workflow",
		Model:        model.Name,
		SystemPrompt: "Return JSON.",
		Prompt: func(input testWorkflowInput) string {
			return "Message: " + input.Message
		},
	}

	svc := s.makeService(ctx, WithDefinedModel(model), WithWorkflow(workflow))
	runner, runnerErr := workflow.GetRunner(svc)
	s.Require().NoError(runnerErr)

	output, runErr := runner.Run(ctx, testWorkflowInput{Message: "hello"})
	s.Require().NoError(runErr)
	s.Require().True(output.OK, "expected output to be ok")
}
