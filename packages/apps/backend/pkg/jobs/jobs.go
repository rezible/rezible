package jobs

import (
	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type Worker[Args river.JobArgs] = river.Worker[Args]

var (
	UniqueStateNonCompleted = []rivertype.JobState{
		rivertype.JobStatePending,
		rivertype.JobStateAvailable,
		rivertype.JobStateScheduled,
		rivertype.JobStateRunning,
		rivertype.JobStateRetryable,
	}
)

type ProjectNormalizedEvent struct {
	EventId uuid.UUID
}

func (ProjectNormalizedEvent) Kind() string {
	return "project-normalized-event"
}

func (ProjectNormalizedEvent) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: UniqueStateNonCompleted,
		},
	}
}

type SyncIntegrationSourceEvents struct {
	IntegrationId uuid.UUID `json:"integration_id"`
	Sources       []string  `json:"sources"`
	SyncReason    string    `json:"sync_reason,omitempty"`
}

func (SyncIntegrationSourceEvents) Kind() string {
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

const AgentTurnsQueue = "agent-turns"

type StartAgentSession struct {
	SessionID uuid.UUID `json:"agent_session_id" river:"unique"`
}

func (StartAgentSession) Kind() string {
	return "start-agent-session"
}

func (StartAgentSession) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       AgentTurnsQueue,
		MaxAttempts: 3,
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: UniqueStateNonCompleted,
		},
	}
}

type InvokeAgentTurn struct {
	AgentSessionID uuid.UUID `json:"agent_session_id"`
	AgentTurnID    uuid.UUID `json:"agent_turn_id" river:"unique"`
}

func (InvokeAgentTurn) Kind() string {
	return "invoke-agent-turn"
}

func (InvokeAgentTurn) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       AgentTurnsQueue,
		MaxAttempts: 3,
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: UniqueStateNonCompleted,
		},
	}
}
