package jobs

import (
	"context"

	"github.com/riverqueue/river"
)

var workers = river.NewWorkers()

func GetWorkers() *river.Workers {
	return workers
}

func RegisterWorker[A river.JobArgs](worker river.Worker[A]) {
	river.AddWorker[A](workers, worker)
}

func RegisterWorkerFunc[A river.JobArgs](work func(ctx context.Context, args A) error) {
	RegisterWorker[A](river.WorkFunc(func(ctx context.Context, j *river.Job[A]) error {
		return work(ctx, j.Args)
	}))
	//if err != nil {
	//	slog.Warn("failed to register worker", "error", err.Error())
	//}
}
