package agents

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type (
	WorkflowState interface {
		Validate() error
	}

	WorkflowOutput interface {
		Encode() ([]byte, error)
	}

	WorkflowDefinition[State WorkflowState, Output WorkflowOutput] struct {
		name string
	}

	AnyWorkflow = WorkflowDefinition[WorkflowState, WorkflowOutput]
)

func (w WorkflowDefinition[S, O]) Name() string {
	return w.name
}

func (w WorkflowDefinition[S, O]) ValidateInput(input []byte) (*S, error) {
	var s S
	if jsonErr := json.Unmarshal(input, &s); jsonErr != nil {
		return nil, jsonErr
	}
	return &s, s.Validate()
}

type workflowValidator struct {
	validateInputFn func(input []byte) error
}

var workflows = map[string]workflowValidator{}

func defineWorkflow[S WorkflowState, O WorkflowOutput](name string) WorkflowDefinition[S, O] {
	w := WorkflowDefinition[S, O]{name: name}
	workflows[name] = workflowValidator{
		validateInputFn: func(input []byte) error {
			_, err := w.ValidateInput(input)
			return err
		},
	}
	return w
}

var WorkflowAlertInvestigation = defineWorkflow[AlertInvestigationState, AlertInvestigationOutput]("alert_investigation")

func ValidateInput(workflowName string, input []byte) error {
	v, ok := workflows[workflowName]
	if !ok {
		return fmt.Errorf("invalid workflow name: %s", workflowName)
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
