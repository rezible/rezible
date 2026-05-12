package jobs

import (
	"github.com/google/uuid"
)

type ProjectNormalizedEvent struct {
	EventId uuid.UUID
}

func (ProjectNormalizedEvent) Kind() string {
	return "project-normalized-event"
}

type ProviderEventSyncJob struct {
	ProviderSources map[string][]string `json:"provider_sources,omitempty"`
	CursorAfter     *string             `json:"cursor_after,omitempty"`
	SyncReason      string              `json:"sync_reason,omitempty"`
}

func (ProviderEventSyncJob) Kind() string {
	return "provider-event-sync"
}

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

type GenerateIncidentDebriefSuggestions struct {
	DebriefId uuid.UUID
}

func (GenerateIncidentDebriefSuggestions) Kind() string {
	return "generate-incident-debrief-suggestions"
}

type ScanOncallShifts struct{}

func (ScanOncallShifts) Kind() string {
	return "scan-oncall-shifts"
}

type EnsureShiftHandoverSent struct {
	ShiftId uuid.UUID
}

func (EnsureShiftHandoverSent) Kind() string { return "ensure-shift-handover-sent" }

type EnsureShiftHandoverReminderSent struct {
	ShiftId uuid.UUID
}

func (EnsureShiftHandoverReminderSent) Kind() string { return "ensure-shift-handover-reminder-sent" }

type GenerateShiftMetrics struct {
	ShiftId uuid.UUID
}

func (GenerateShiftMetrics) Kind() string {
	return "generate-shift-metrics"
}
