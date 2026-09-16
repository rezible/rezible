package jobs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/pkg/execution"

	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
)

// SetContextualArgs populates context-derived fields before River encodes arguments and calculates uniqueness. Ordinary arguments are unchanged.
func SetContextualArgs(ctx context.Context, args JobArgs) error {
	if setter, hasTenantID := args.(tenantIDSetter); hasTenantID {
		tenantID, tenantOK := execution.GetContext(ctx).TenantID()
		if !tenantOK {
			return rez.ErrTenantContextMissing
		}
		setter.setTenantID(tenantID)
	}
	return nil
}

type tenantIDSetter interface{ setTenantID(int) }

// TenantArgs scopes job identity to the enqueueing tenant. Embed it by value
// and define Kind on a pointer receiver so context fields can be populated.
// Other unique fields must also carry river:"unique" tags.
type TenantArgs struct {
	TenantID int `json:"tenant_id" river:"unique"`
}

func (a *TenantArgs) setTenantID(tenantID int) {
	a.TenantID = tenantID
}

type ProcessProviderEventArgs struct {
	TenantArgs
	Event rez.ProviderEvent `json:"event" river:"unique"`
}

func (*ProcessProviderEventArgs) Kind() string {
	return "process-provider-event"
}

func (*ProcessProviderEventArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}}
}

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

type BumpSituationInvestigation struct {
	SituationInvestigationID uuid.UUID `json:"situation_investigation_id"`
}

func (BumpSituationInvestigation) Kind() string {
	return "bump-situation-investigation"
}

type CloseInactiveAlertEpisodes struct{}

func (CloseInactiveAlertEpisodes) Kind() string {
	return "close-inactive-alert-episodes"
}
func (CloseInactiveAlertEpisodes) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: UniqueStateNonCompleted,
		},
	}
}

var CloseInactiveAlertEpisodesPeriodicJob = river.NewPeriodicJob(
	river.PeriodicInterval(time.Minute),
	func() (river.JobArgs, *river.InsertOpts) {
		return CloseInactiveAlertEpisodes{}, nil
	},
	nil,
)
