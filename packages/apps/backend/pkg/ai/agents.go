package agents

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type (
	WorkflowInput interface {
		Validate() error
	}

	WorkflowOutput interface {
		Encode() ([]byte, error)
	}

	WorkflowDefinition[I WorkflowInput, Output WorkflowOutput] struct {
		name string
	}
)

func (w WorkflowDefinition[I, O]) ValidateJSONInput(input []byte) (*I, error) {
	var i I
	if jsonErr := json.Unmarshal(input, &i); jsonErr != nil {
		return nil, jsonErr
	}
	return &i, i.Validate()
}

type (
	AgentState interface {
		Validate() error
	}

	AgentDefinition[State AgentState] struct {
		name string
	}
)

func (a AgentDefinition[S]) ValidateJSONInput(input []byte) (*S, error) {
	var s S
	if jsonErr := json.Unmarshal(input, &s); jsonErr != nil {
		return nil, jsonErr
	}
	return &s, s.Validate()
}

type (
	jsonInputValidator[S any] interface {
		ValidateJSONInput([]byte) (*S, error)
	}

	inputValidator struct {
		validateInputFn func(input any) error
	}
)

func makeInputValidator[S any](a jsonInputValidator[S]) inputValidator {
	return inputValidator{
		validateInputFn: func(input any) error {
			enc, ok := input.([]byte)
			if !ok {
				var marshalErr error
				enc, marshalErr = json.Marshal(input)
				if marshalErr != nil {
					return fmt.Errorf("marshal input: %w", marshalErr)
				}
			}
			_, validationErr := a.ValidateJSONInput(enc)
			return validationErr
		},
	}
}

var agents = map[string]inputValidator{}

func defineAgent[S AgentState](name string) AgentDefinition[S] {
	a := AgentDefinition[S]{name: name}
	agents[name] = makeInputValidator(a)
	return a
}

var AgentAlertInvestigation = defineAgent[AlertInvestigationState]("alert_investigation")

func ValidateAgentInput(name string, input any) error {
	v, ok := agents[name]
	if !ok {
		return fmt.Errorf("invalid workflow name: %s", name)
	}
	return v.validateInputFn(input)
}

type (
	AlertInvestigationState struct {
		AlertID uuid.UUID `json:"alert_id"`
	}

	AlertInvestigationOutput struct {
		Limitations        []string                   `json:"limitations"`
		RecommendedActions []string                   `json:"recommendedActions"`
		Findings           AlertInvestigationFindings `json:"findings"`
	}

	AlertInvestigationFindings struct {
		LikelyCause     string   `json:"likelyCause"`
		AffectedSystems []string `json:"affectedSystems"`
		SuggestedChecks []string `json:"suggestedChecks"`
		RecommendedNext string   `json:"recommendedNext"`
	}
)

func (i AlertInvestigationState) Validate() error {
	return nil
}

func (i AlertInvestigationOutput) Encode() ([]byte, error) {
	return json.Marshal(i)
}
