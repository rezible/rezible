package schematypes

// InvestigationNextCheck is something a responder can look at next, suggested with an investigation report.
type InvestigationNextCheck struct {
	Check  string `json:"check"`
	Reason string `json:"reason"`
}
