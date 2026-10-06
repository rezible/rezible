package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/pkg/execution"

	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
)

// ArgsWithWorkerExecutionContext lets a job's args choose the identity its worker runs with,
// instead of the default tenant actor.
type ArgsWithWorkerExecutionContext interface {
	WorkerExecutionContext() execution.ActorKind
}

// SetContextualArgs populates context-derived fields before River encodes arguments and calculates uniqueness. Ordinary arguments are unchanged.
func SetContextualArgs(ctx context.Context, args JobArgs) error {
	if tenantArgs, hasTenantID := args.(tenantIDArgs); hasTenantID {
		tenantID, tenantOK := execution.GetContext(ctx).TenantID()
		if !tenantOK {
			return rez.ErrTenantContextMissing
		}
		if current := tenantArgs.tenantID(); current != 0 && current != tenantID {
			return fmt.Errorf("%w: job tenant does not match the inserting context", rez.ErrForbidden)
		}
		tenantArgs.setTenantID(tenantID)
	}
	return nil
}

type tenantIDArgs interface {
	tenantID() int
	setTenantID(int)
}

// TenantArgs scopes job uniqueness to the enqueueing tenant; it does not affect
// the worker's identity. Embed it by value and define Kind on a pointer receiver
// so context fields can be populated. Other unique fields must also carry
// river:"unique" tags.
type TenantArgs struct {
	TenantID int `json:"tenant_id" river:"unique"`
}

func (a *TenantArgs) tenantID() int {
	return a.TenantID
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

type ReconcileInvestigation struct {
	InvestigationID uuid.UUID `json:"investigation_id"`
}

func (ReconcileInvestigation) Kind() string {
	return "reconcile-investigation"
}

// SettleAlertEpisode recomputes an alert episode's state at a deadline: a window timeout, a supersession
// or the episode's closing time. It is unique by args, so repeated scheduling of one deadline is one job.
type SettleAlertEpisode struct {
	EpisodeID uuid.UUID `json:"episode_id"`
	DueAt     time.Time `json:"due_at"`
}

func (SettleAlertEpisode) Kind() string {
	return "settle-alert-episode"
}

func (SettleAlertEpisode) InsertOpts() river.InsertOpts {
	return river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}}
}

// ProcessSituationSignal places a changed signal into situations, or refreshes the situation holding it.
// It is not unique: a trigger dropped while an identical job runs would miss the change, so its handler is
// idempotent instead.
type ProcessSituationSignal struct {
	SignalEntityID uuid.UUID `json:"signal_entity_id"`
}

func (ProcessSituationSignal) Kind() string {
	return "process-situation-signal"
}

// SituationsQueue runs situation evaluations, which may wait on a model judge, apart from the default queue.
const SituationsQueue = "situations"

// EvaluateSituation brings a situation's closure, investigation and judgment up to date. An immediate
// trigger has no due time and is not unique: a trigger dropped while an identical job runs would miss the
// change, so its handler is idempotent instead. A deadline carries its due time and is unique by args, so
// repeated scheduling of one deadline is one job.
type EvaluateSituation struct {
	SituationID uuid.UUID  `json:"situation_id"`
	DueAt       *time.Time `json:"due_at,omitempty"`
}

func (EvaluateSituation) Kind() string {
	return "evaluate-situation"
}

func (a EvaluateSituation) InsertOpts() river.InsertOpts {
	if a.DueAt == nil {
		return river.InsertOpts{Queue: SituationsQueue}
	}
	return river.InsertOpts{Queue: SituationsQueue, UniqueOpts: river.UniqueOpts{ByArgs: true}}
}
