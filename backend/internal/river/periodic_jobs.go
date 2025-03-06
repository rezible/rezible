package river

import (
	"context"
	"errors"
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

func (s *JobService) registerProviderDataSyncPeriodicJob(
	interval time.Duration,
	users rez.UserService,
	teams rez.TeamService,
	incidents rez.IncidentService,
	oncall rez.OncallService,
	alerts rez.AlertsService,
	components rez.SystemComponentsService,
) error {
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

	workFn := river.WorkFunc(func(ctx context.Context, j *river.Job[jobs.SyncProviderData]) error {
		var err error
		if j.Args.Users {
			err = errors.Join(err, users.SyncData(ctx))
		}
		if j.Args.Teams {
			err = errors.Join(err, teams.SyncData(ctx))
		}
		if j.Args.Oncall {
			err = errors.Join(err, oncall.SyncData(ctx))
		}
		if j.Args.Incidents {
			err = errors.Join(err, incidents.SyncData(ctx))
		}
		if j.Args.Alerts {
			err = errors.Join(err, alerts.SyncData(ctx))
		}
		if j.Args.SystemComponents {
			err = errors.Join(err, components.SyncData(ctx))
		}
		return err
	})

	return river.AddWorkerSafely(s.clientCfg.Workers, workFn)
}
