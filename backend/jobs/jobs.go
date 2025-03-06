package jobs

import "github.com/google/uuid"

type (
	JobArgs interface {
		Kind() string
	}

	InsertOpts struct {
		Uniqueness *UniquenessOpts
	}

	UniquenessOpts struct {
		Args bool
	}

	InsertManyParams struct {
		Args JobArgs
		Opts *InsertOpts
	}
)

type SyncProviderData struct {
	Users            bool
	Teams            bool
	Incidents        bool
	Oncall           bool
	Alerts           bool
	SystemComponents bool
}

func (SyncProviderData) Kind() string {
	return "sync-provider-data"
}

// SendIncidentDebriefRequests Send requests for users to complete debriefs
type SendIncidentDebriefRequests struct {
	IncidentId uuid.UUID
}

func (SendIncidentDebriefRequests) Kind() string {
	return "send-incident-debrief-requests"
}

type GenerateIncidentDebriefResponse struct {
	DebriefId uuid.UUID
}

func (GenerateIncidentDebriefResponse) Kind() string {
	return "generate-incident-debrief-response"
}

// GenerateIncidentDebriefSuggestions Generate Debrief Suggestions
type GenerateIncidentDebriefSuggestions struct {
	DebriefId uuid.UUID
}

func (GenerateIncidentDebriefSuggestions) Kind() string {
	return "generate-incident-debrief-suggestions"
}

type ScanOncallHandovers struct{}

func (ScanOncallHandovers) Kind() string {
	return "scan-oncall-handovers"
}

// EnsureShiftHandover (send user reminder, or auto-send fallback if unsent)
type EnsureShiftHandover struct {
	ShiftId uuid.UUID
}

func (EnsureShiftHandover) Kind() string {
	return "send-shift-handover-reminder"
}
