package ai

import (
	"github.com/google/uuid"
)

var AlertInvestigationAgent = defineAgent[AlertInvestigationState]("alert_investigation")

type (
	AlertInvestigationState struct {
		AlertID uuid.UUID `json:"alert_id"`
	}

	//AlertInvestigationOutput struct {
	//	Limitations        []string                   `json:"limitations"`
	//	RecommendedActions []string                   `json:"recommendedActions"`
	//	Findings           AlertInvestigationFindings `json:"findings"`
	//}
	//
	//AlertInvestigationFindings struct {
	//	LikelyCause     string   `json:"likelyCause"`
	//	AffectedSystems []string `json:"affectedSystems"`
	//	SuggestedChecks []string `json:"suggestedChecks"`
	//	RecommendedNext string   `json:"recommendedNext"`
	//}
)

func (i AlertInvestigationState) Validate() error {
	return nil
}
