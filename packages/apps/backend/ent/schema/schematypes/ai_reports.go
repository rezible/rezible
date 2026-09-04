package schematypes

type (
	SituationInvestigationReport struct {
		Text               string   `json:"text"`
		Limitations        []string `json:"limitations,omitempty"`
		LikelyCause        string   `json:"likely_cause,omitempty"`
		RecommendedActions []string `json:"recommended_actions,omitempty"`
		SuggestedChecks    []string `json:"suggested_checks,omitempty"`
		BestNextStep       string   `json:"best_next_step,omitempty"`
	}
)
