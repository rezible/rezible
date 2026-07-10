package ai

import (
	"encoding/json"
	"fmt"
)

type (
	WorkflowInput interface {
		Validate() error
	}

	WorkflowState interface{}

	WorkflowOutput interface {
		Encode() ([]byte, error)
	}

	WorkflowDefinition[I WorkflowInput, O WorkflowOutput, S WorkflowState] struct {
		Name   string
		Prompt *string
	}
)

func (w WorkflowDefinition[I, O, S]) ValidateInput(input []byte) (*I, error) {
	var i I
	if jsonErr := json.Unmarshal(input, &i); jsonErr != nil {
		return nil, jsonErr
	}
	return &i, i.Validate()
}

type workflowValidator struct {
	validateInputFn  func(input any) ([]byte, error)
	validateOutputFn func(output any) ([]byte, error)
}

var workflows = map[string]workflowValidator{}

func makeWorkflowValidator[I WorkflowInput, O WorkflowOutput, S WorkflowState](w WorkflowDefinition[I, O, S]) workflowValidator {
	return workflowValidator{
		validateInputFn: func(input any) ([]byte, error) {
			enc, ok := input.([]byte)
			if !ok {
				var marshalErr error
				enc, marshalErr = json.Marshal(input)
				if marshalErr != nil {
					return nil, fmt.Errorf("marshal input: %w", marshalErr)
				}
			}
			_, validationErr := w.ValidateInput(enc)
			return enc, validationErr
		},
	}
}

func defineWorkflow[I WorkflowInput, O WorkflowOutput, S WorkflowState](name string) WorkflowDefinition[I, O, S] {
	w := WorkflowDefinition[I, O, S]{Name: name}
	workflows[name] = makeWorkflowValidator(w)
	return w
}

func ValidateWorkflowInput(name string, input any) ([]byte, error) {
	v, ok := workflows[name]
	if !ok {
		return nil, fmt.Errorf("invalid workflow name: %s", name)
	}
	return v.validateInputFn(input)
}

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

var ExampleWorkflow = defineWorkflow[ExampleWorkflowInput, ExampleWorkflowOutput, ExampleWorkflowState]("example")
