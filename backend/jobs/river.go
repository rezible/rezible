package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/riverqueue/river"
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

	WorkFn[T JobArgs] = func(ctx context.Context, args T) error

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

	riverWorkerRegistry *river.Workers
)

func SetWorkerRegistry(reg any) error {
	switch typedReg := reg.(type) {
	case *river.Workers:
		riverWorkerRegistry = typedReg
	default:
		return fmt.Errorf("unknown worker registry type")
	}
	return nil
}

func RegisterWorkerFunc[A JobArgs](worker WorkFn[A]) error {
	if riverWorkerRegistry != nil {
		workFn := river.WorkFunc(func(ctx context.Context, j *river.Job[A]) error {
			return worker(ctx, j.Args)
		})
		return river.AddWorkerSafely[A](riverWorkerRegistry, workFn)
	}
	return fmt.Errorf("no worker registry set")
}

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
