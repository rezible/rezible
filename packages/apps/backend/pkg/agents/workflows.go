package agents

import "github.com/google/uuid"

type Workflow[State any, Output any] struct {
	name        string
	description string
}

func (w Workflow[S, O]) Name() string {
	return w.name
}

func (w Workflow[S, O]) Description() string {
	return w.description
}

func NewWorkflow[S any, O any](name string, description string) Workflow[S, O] {
	return Workflow[S, O]{name: name, description: description}
}

var WorkflowAlertInvestigation = NewWorkflow[AlertInvestigationState, AlertInvestigationOutput]("alert_investigation", "TODO")

type (
	AlertInvestigationState struct {
		AlertID uuid.UUID
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
