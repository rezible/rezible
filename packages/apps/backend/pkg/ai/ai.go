package ai

import (
	"encoding/json"
	"fmt"

	"github.com/firebase/genkit/go/ai"
)

type (
	Message = ai.Message
)

type (
	AgentState interface {
		Validate() error
	}

	AgentDefinition[State AgentState] struct {
		Name string
	}
)

func (a AgentDefinition[S]) CreateInitialState(input []byte) (*S, error) {
	var s S
	if jsonErr := json.Unmarshal(input, &s); jsonErr != nil {
		return nil, jsonErr
	}
	return &s, s.Validate()
}

type (
	agentInputValidatorFn func(input any) ([]byte, error)

	agentValidator struct {
		validateInput agentInputValidatorFn
	}
)

func makeAgentValidator[S AgentState](a AgentDefinition[S]) agentValidator {
	return agentValidator{
		validateInput: func(input any) ([]byte, error) {
			enc, ok := input.([]byte)
			if !ok {
				var marshalErr error
				enc, marshalErr = json.Marshal(input)
				if marshalErr != nil {
					return nil, fmt.Errorf("marshal input: %w", marshalErr)
				}
			}
			_, validationErr := a.CreateInitialState(enc)
			return enc, validationErr
		},
	}
}

var agents = map[string]agentValidator{}

func ValidateAndEncodeAgentInput(name string, input any) ([]byte, error) {
	v, ok := agents[name]
	if !ok {
		return nil, fmt.Errorf("invalid agent name: %s", name)
	}
	return v.validateInput(input)
}

func defineAgent[S AgentState](name string) AgentDefinition[S] {
	a := AgentDefinition[S]{Name: name}
	agents[name] = makeAgentValidator(a)
	return a
}

type (
	WorkflowInput interface {
		Validate() error
	}

	WorkflowOutput interface {
		Encode() ([]byte, error)
	}

	WorkflowDefinition[I WorkflowInput, S any, O WorkflowOutput] struct {
		Name string
	}
)

func (w WorkflowDefinition[I, S, O]) ValidateInput(input []byte) (*I, error) {
	var i I
	if jsonErr := json.Unmarshal(input, &i); jsonErr != nil {
		return nil, jsonErr
	}
	return &i, i.Validate()
}

type (
	workflowInputValidatorFn func(input any) ([]byte, error)

	workflowValidator struct {
		validateInputFn  workflowInputValidatorFn
		validateOutputFn func(output any) ([]byte, error)
	}
)

func makeWorkflowValidator[I WorkflowInput, S any, O WorkflowOutput](w WorkflowDefinition[I, S, O]) workflowValidator {
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

var workflows = map[string]workflowValidator{}

func ValidateWorkflowInput(name string, input any) ([]byte, error) {
	v, ok := workflows[name]
	if !ok {
		return nil, fmt.Errorf("invalid workflow name: %s", name)
	}
	return v.validateInputFn(input)
}

func defineWorkflow[I WorkflowInput, S any, O WorkflowOutput](name string) WorkflowDefinition[I, S, O] {
	w := WorkflowDefinition[I, S, O]{Name: name}
	workflows[name] = makeWorkflowValidator(w)
	return w
}
