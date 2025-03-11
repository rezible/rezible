package river

import (
	"context"
	"time"

	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/jobs"
)

func (s *JobService) registerOncallHandoverScanPeriodicJob(interval time.Duration, oncall rez.OncallService) error {
	s.addPeriodicJob(river.NewPeriodicJob(
		river.PeriodicInterval(interval),
		func() (river.JobArgs, *river.InsertOpts) {
			return &jobs.ScanOncallHandovers{}, nil
		},
		&river.PeriodicJobOpts{
			RunOnStart: true,
		},
	))

	worker := river.WorkFunc(func(ctx context.Context, j *river.Job[jobs.ScanOncallHandovers]) error {
		return oncall.ScanForShiftsNeedingHandover(ctx)
	})

	return river.AddWorkerSafely(s.clientCfg.Workers, worker)
}

func (s *JobService) RegisterProviderDataSyncPeriodicJob(interval time.Duration, workFn func(ctx context.Context, args jobs.SyncProviderData) error) error {
	args := &jobs.SyncProviderData{
		Users:            true,
		Teams:            true,
		Incidents:        true,
		Oncall:           true,
		Alerts:           true,
		SystemComponents: true,
	}
	opts := &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByState: NonCompletedJobStates,
		},
	}
	s.addPeriodicJob(river.NewPeriodicJob(
		river.PeriodicInterval(interval),
		func() (river.JobArgs, *river.InsertOpts) {
			return args, opts
		},
		&river.PeriodicJobOpts{
			RunOnStart: true,
		},
	))

	return river.AddWorkerSafely(s.clientCfg.Workers, river.WorkFunc(func(ctx context.Context, j *river.Job[jobs.SyncProviderData]) error {
		return workFn(ctx, j.Args)
	}))
}
