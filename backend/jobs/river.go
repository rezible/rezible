package jobs

import (
	"context"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type (
	JobArgs          = river.JobArgs
	InsertOpts       = river.InsertOpts
	InsertManyParams = river.InsertManyParams
	UniqueOpts       = river.UniqueOpts

	Job[T JobArgs]            = river.Job[T]
	Worker[T JobArgs]         = river.Worker[T]
	WorkerDefaults[T JobArgs] = river.WorkerDefaults[T]
	WorkFn[T JobArgs]         = func(ctx context.Context, args T) error

	JobState = rivertype.JobState

	PeriodicJob            = river.PeriodicJob
	PeriodicJobConstructor = river.PeriodicJobConstructor
	PeriodicJobOpts        = river.PeriodicJobOpts
	PeriodicSchedule       = river.PeriodicSchedule
)

var (
	Workers      = river.NewWorkers()
	PeriodicJobs []*PeriodicJob
)

func RegisterPeriodicJob(job *PeriodicJob) {
	PeriodicJobs = append(PeriodicJobs, job)
}

func PeriodicInterval(interval time.Duration) PeriodicSchedule {
	return river.PeriodicInterval(interval)
}

func NewPeriodicJob(sched PeriodicSchedule, cs PeriodicJobConstructor, opts *PeriodicJobOpts) *PeriodicJob {
	return river.NewPeriodicJob(sched, cs, opts)
}

func RegisterWorker[A JobArgs](worker Worker[A]) {
	river.AddWorker[A](Workers, worker)
}

func RegisterWorkerFunc[A JobArgs](work WorkFn[A]) {
	river.AddWorker[A](Workers, river.WorkFunc(func(ctx context.Context, j *river.Job[A]) error {
		return work(ctx, j.Args)
	}))
}

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
