package jobs

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/riverqueue/river"
)

type (
	JobRegistrar interface {
		RegisterJobs(*Registry) error
	}
	WorkerRegistrar interface {
		RegisterWorkers(*Registry) error
	}
)

type Registry struct {
	mu           sync.Mutex
	workers      *river.Workers
	periodicJobs []*river.PeriodicJob
	kinds        map[string]struct{}
	sealed       bool
}

type Definition struct {
	Workers      *river.Workers
	PeriodicJobs []*river.PeriodicJob
	Kinds        []string
}

func NewRegistry() *Registry {
	return &Registry{
		workers: river.NewWorkers(),
		kinds:   make(map[string]struct{}),
	}
}

func (r *Registry) AddWorker[A river.JobArgs](worker river.Worker[A]) error {
	if r == nil {
		return fmt.Errorf("job registry is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var args A
	kind := args.Kind()
	if r.sealed {
		return fmt.Errorf("register worker %q: job registry is sealed", kind)
	}
	if addErr := river.AddWorkerSafely(r.workers, worker); addErr != nil {
		return fmt.Errorf("register worker %q: %w", kind, addErr)
	}
	r.kinds[kind] = struct{}{}
	return nil
}

func (r *Registry) AddWorkerFunc[A river.JobArgs](work func(context.Context, A) error) error {
	worker := river.WorkFunc(func(ctx context.Context, job *river.Job[A]) error {
		return work(ctx, job.Args)
	})
	return r.AddWorker(worker)
}

func (r *Registry) AddPeriodicJob(job *river.PeriodicJob) error {
	if r == nil {
		return fmt.Errorf("job registry is nil")
	}
	if job == nil {
		return fmt.Errorf("periodic job is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sealed {
		return fmt.Errorf("register periodic job: job registry is sealed")
	}
	r.periodicJobs = append(r.periodicJobs, job)
	return nil
}

func (r *Registry) Kinds() []string {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.sortedKinds()
}

func (r *Registry) Seal() (*Definition, error) {
	if r == nil {
		return nil, fmt.Errorf("job registry is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sealed {
		return nil, fmt.Errorf("job registry is already sealed")
	}
	r.sealed = true
	return &Definition{
		Workers:      r.workers,
		PeriodicJobs: append([]*river.PeriodicJob(nil), r.periodicJobs...),
		Kinds:        r.sortedKinds(),
	}, nil
}

func (r *Registry) sortedKinds() []string {
	kinds := make([]string, 0, len(r.kinds))
	for kind := range r.kinds {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return kinds
}
