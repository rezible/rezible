package jobs

import (
	"github.com/google/uuid"
	"time"
)

type (
	JobArgs interface {
		Kind() string
	}

	InsertManyParams struct {
		Args JobArgs
		Opts *InsertOpts
	}

	InsertOpts struct {
		Uniqueness *UniquenessOpts
	}

	UniquenessOpts struct {
		Args        bool
		ByState     []JobState
		ByPeriod    time.Duration
		ByQueue     bool
		ExcludeKind bool
	}

	JobState string
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

// see river.PeriodicJob
type (
	PeriodicJob struct {
		ConstructorFunc PeriodicJobConstructor
		Opts            *PeriodicJobOpts
		Schedule        PeriodicSchedule
	}
	PeriodicSchedule interface {
		Next(current time.Time) time.Time
	}
	PeriodicJobConstructor func() (JobArgs, *InsertOpts)
	PeriodicJobOpts        struct {
		RunOnStart bool
	}
	periodicIntervalSchedule struct {
		interval time.Duration
	}
)

func NewPeriodicJob(scheduleFunc PeriodicSchedule, constructorFunc PeriodicJobConstructor, opts *PeriodicJobOpts) *PeriodicJob {
	return &PeriodicJob{
		ConstructorFunc: constructorFunc,
		Opts:            opts,
		Schedule:        scheduleFunc,
	}
}

func PeriodicInterval(interval time.Duration) PeriodicSchedule {
	return &periodicIntervalSchedule{interval}
}
func (s *periodicIntervalSchedule) Next(t time.Time) time.Time {
	return t.Add(s.interval)
}

type SyncProviderData struct {
	Hard bool

	Users            bool
	Teams            bool
	Incidents        bool
	Oncall           bool
	OncallEvents     bool
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
