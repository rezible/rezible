package jobs

import (
	"context"
	"fmt"

	"github.com/riverqueue/river"
)

type (
	WorkFn[T JobArgs] = func(ctx context.Context, args T) error
)

var (
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
