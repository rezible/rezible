package schematypes

import "time"

// InvestigationNextCheck is something a responder can look at next, suggested with an investigation report.
type InvestigationNextCheck struct {
	Check  string `json:"check"`
	Reason string `json:"reason"`
}

// ProviderReadItem is one citable item a provider read returned: a log line with its text, or a counted point
// with its value.
type ProviderReadItem struct {
	// Ref is the item's ref within the read, 1 to N in the order returned.
	Ref string `json:"ref"`
	// Time is the newest repeat's time when Count is above one.
	Time  time.Time `json:"time"`
	Text  string    `json:"text,omitempty"`
	Value *float64  `json:"value,omitempty"`
	// Labels are the item's labels; in a search result, those not shared by every item.
	Labels map[string]string `json:"labels,omitempty"`
	// Count is how many lines this entry stands for, identical or differing only in identifiers; omitted when one.
	Count int `json:"count,omitempty"`
	// Earliest is the oldest repeat's time; omitted when Count is one.
	Earliest *time.Time `json:"earliest,omitempty"`
}
