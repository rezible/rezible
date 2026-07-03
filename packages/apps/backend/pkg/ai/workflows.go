package ai

import (
	"encoding/json"
)

type (
	ExampleWorkflowInput struct {
		Name string `json:"name"`
	}

	ExampleWorkflowState struct{}

	ExampleWorkflowOutput struct {
		Message string `json:"message"`
	}
)

func (i ExampleWorkflowInput) Validate() error {
	return nil
}

func (o ExampleWorkflowOutput) Encode() ([]byte, error) {
	return json.Marshal(o)
}

var ExampleWorkflow = defineWorkflow[ExampleWorkflowInput, ExampleWorkflowState, ExampleWorkflowOutput]("example")
