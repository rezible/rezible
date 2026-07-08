package jobs

import (
	"github.com/firebase/genkit/go/ai"
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

type PopulateKnowledgeGraphSnapshot struct {
	SnapshotId uuid.UUID `json:"snapshot_id"`
}

func (PopulateKnowledgeGraphSnapshot) Kind() string {
	return "create-knowledge-graph-snapshot"
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

type StartAgentRun struct {
	AgentRunID uuid.UUID `json:"agent_run_id"`
}

func (StartAgentRun) Kind() string {
	return "start-agent-run"
}

type ContinueAgentRun struct {
	AgentRunID       uuid.UUID                `json:"agent_run_id"`
	ParentSnapshotID *uuid.UUID               `json:"parent_snapshot_id,omitempty"`
	Message          *ai.Message              `json:"message,omitempty"`
	Resume           *ai.GenerateActionResume `json:"resume,omitempty"`
}

func (ContinueAgentRun) Kind() string {
	return "invoke-agent-run"
}
