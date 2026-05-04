package jobs

import (
	"context"

	"github.com/riverqueue/river"
)

var (
	workers      = river.NewWorkers()
	periodicJobs []*river.PeriodicJob
)

func GetWorkers() *river.Workers {
	return workers
}

func GetPeriodicJobs() []*river.PeriodicJob {
	return periodicJobs
}

func RegisterPeriodicJob(job *river.PeriodicJob) {
	periodicJobs = append(periodicJobs, job)
}

func RegisterWorker[A river.JobArgs](worker river.Worker[A]) {
	river.AddWorker[A](workers, worker)
}

func RegisterWorkerFunc[A river.JobArgs](work func(ctx context.Context, args A) error) {
	river.AddWorker[A](workers, river.WorkFunc(func(ctx context.Context, j *river.Job[A]) error {
		return work(ctx, j.Args)
	}))
}
