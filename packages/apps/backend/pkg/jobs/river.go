package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type (
	JobArgs                   = river.JobArgs
	Job[A JobArgs]            = river.Job[A]
	PeriodicJob               = river.PeriodicJob
	Worker[A JobArgs]         = river.Worker[A]
	WorkerDefaults[A JobArgs] = river.WorkerDefaults[A]

	Definition struct {
		Workers      []WorkerDefinition
		PeriodicJobs []*PeriodicJob
	}
)

var (
	UniqueStateNonCompleted = []rivertype.JobState{
		rivertype.JobStatePending,
		rivertype.JobStateAvailable,
		rivertype.JobStateScheduled,
		rivertype.JobStateRunning,
		rivertype.JobStateRetryable,
	}
)

type WorkerDefinition struct {
	kind     string
	register func(*river.Workers) error
}

func (d WorkerDefinition) Register(w *river.Workers) error {
	return d.register(w)
}

func (d WorkerDefinition) Kind() string {
	return d.kind
}

func DefineWorker[A river.JobArgs](worker river.Worker[A]) WorkerDefinition {
	var args A
	return WorkerDefinition{
		kind: args.Kind(),
		register: func(workers *river.Workers) error {
			return river.AddWorkerSafely(workers, worker)
		},
	}
}

type WorkerFunc[A JobArgs] func(context.Context, *Job[A]) error

func DefineWorkerFunc[A river.JobArgs](work func(context.Context, A) error) WorkerDefinition {
	return DefineWorker[A](river.WorkFunc(func(ctx context.Context, job *river.Job[A]) error {
		return work(ctx, job.Args)
	}))
}

var ErrRetryable = errors.New("retryable job failure")

// DefaultWorkerTimeout is a sane default for workers that do not derive their
// timeout from configuration.
const DefaultWorkerTimeout = 15 * time.Minute

func MarkRetryableError(err error) error {
	if err == nil {
		return ErrRetryable
	}
	return fmt.Errorf("%w: %w", ErrRetryable, err)
}

func IsRetryableError(err error) bool {
	return errors.Is(err, ErrRetryable)
}
