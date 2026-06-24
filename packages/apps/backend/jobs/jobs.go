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

type SyncIntegrationEventsArgs struct {
	IntegrationId uuid.UUID `json:"integration_id"`
	Sources       []string  `json:"sources"`
	SyncReason    string    `json:"sync_reason,omitempty"`
}

func (SyncIntegrationEventsArgs) Kind() string {
	return "sync-integration-events"
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

type RunAgentWorkflow struct {
	AgentRunID uuid.UUID `json:"agent_run_id"`
}

func (RunAgentWorkflow) Kind() string {
	return "run-agent-workflow"
}

type InvestigateAlert struct {
	AlertID         uuid.UUID `json:"alert_id"`
	AlertInstanceID uuid.UUID `json:"alert_instance_id,omitempty"`
}

func (InvestigateAlert) Kind() string {
	return "investigate-alert"
}

type PostAlertInvestigationUpdate struct {
	AlertID    uuid.UUID `json:"alert_id"`
	AgentRunID uuid.UUID `json:"agent_run_id"`
}

func (PostAlertInvestigationUpdate) Kind() string {
	return "post-alert-investigation-update"
}

type AnalyzeIncidentRetrospective struct {
	IncidentID      uuid.UUID `json:"incident_id"`
	RetrospectiveID uuid.UUID `json:"retrospective_id,omitempty"`
}

func (AnalyzeIncidentRetrospective) Kind() string {
	return "analyze-incident-retrospective"
}
