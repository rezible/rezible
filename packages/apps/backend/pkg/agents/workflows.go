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

func defineWorkflow[State any, Output any](name string) Workflow[State, Output] {
	return Workflow[State, Output]{name: name, description: "TODO"}
}

var WorkflowAlertInvestigation = defineWorkflow[AlertInvestigationState, AlertInvestigationOutput]("alert_investigation")

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
