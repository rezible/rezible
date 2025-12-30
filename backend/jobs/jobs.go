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

type SyncIntegrationsData struct {
	Hard           bool
	CreateDefaults bool
	OrganizationId uuid.UUID
	IntegrationId  uuid.UUID
}

func (SyncIntegrationsData) Kind() string {
	return "sync-integrations-data"
}

var SyncAllTenantIntegrationsDataPeriodicJob = PeriodicJob{
	ConstructorFunc: func() InsertJobParams {
		return InsertJobParams{
			Args: &SyncIntegrationsData{},
			Uniqueness: &JobUniquenessOpts{
				ByState: NonCompletedJobStates,
			},
		}
	},
	Interval: time.Hour,
	Opts:     &PeriodicJobOpts{RunOnStart: true},
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

var ScanOncallShiftsPeriodicJob = PeriodicJob{
	ConstructorFunc: func() InsertJobParams {
		return InsertJobParams{Args: &ScanOncallShifts{}}
	},
	Interval: time.Hour,
	Opts:     &PeriodicJobOpts{RunOnStart: true},
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

type HandleIncidentChatUpdate struct {
	IncidentId uuid.UUID
}

func (HandleIncidentChatUpdate) Kind() string { return "handle-incident-chat-update" }
