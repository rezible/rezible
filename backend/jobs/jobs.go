package jobs

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type (
	JobArgs interface {
		Kind() string
	}

	InsertJobParams struct {
		Args        JobArgs
		ScheduledAt time.Time
		Uniqueness  *JobUniquenessOpts
	}

	JobUniquenessOpts struct {
		Args        bool
		ByState     []JobState
		ByPeriod    time.Duration
		ByQueue     bool
		ExcludeKind bool
	}

	JobState string

	WorkFn[T JobArgs] = func(ctx context.Context, args T) error

	PeriodicJob struct {
		ConstructorFunc func() InsertJobParams
		Interval        time.Duration
		Opts            *PeriodicJobOpts
	}
	PeriodicJobOpts struct {
		RunOnStart bool
	}
)

// see rivertype.JobState
const (
	JobStateAvailable JobState = "available"
	JobStateCancelled JobState = "cancelled"
	JobStateCompleted JobState = "completed"
	JobStateDiscarded JobState = "discarded"
	JobStatePending   JobState = "pending"
	JobStateRetryable JobState = "retryable"
	JobStateRunning   JobState = "running"
	JobStateScheduled JobState = "scheduled"
)

var (
	NonCompletedJobStates = []JobState{JobStateAvailable, JobStatePending, JobStateRunning, JobStateRetryable, JobStateScheduled}
)

type SyncProviderData struct {
	Hard bool
}

func (SyncProviderData) Kind() string {
	return "sync-provider-data"
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

type ProcessChatEvent struct {
	Provider  string
	EventKind string
	Data      any
}

func (ProcessChatEvent) Kind() string {
	return "process-chat-event"
}

type IncidentChatUpdate struct {
	IncidentId      uuid.UUID
	Created         bool
	OriginChannelId string
}

func (IncidentChatUpdate) Kind() string {
	return "incident-chat-update"
}
