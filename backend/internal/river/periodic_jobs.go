package river

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/riverqueue/river"
)

type ensureShiftHandoverJobArgs struct {
	shiftId uuid.UUID
}

func (ensureShiftHandoverJobArgs) Kind() string {
	return "send-shift-handover-reminder"
}

type scanOncallHandoversJobArgs struct{}

func (scanOncallHandoversJobArgs) Kind() string {
	return "scan-oncall-handovers"
}

func (s *JobService) registerOncallHandoverScanPeriodicJob(interval time.Duration, oncall rez.OncallService) error {
	s.addPeriodicJob(river.NewPeriodicJob(
		river.PeriodicInterval(interval),
		func() (river.JobArgs, *river.InsertOpts) {
			return &scanOncallHandoversJobArgs{}, nil
		},
		&river.PeriodicJobOpts{
			RunOnStart: true,
		},
	))

	worker := river.WorkFunc(func(ctx context.Context, j *river.Job[scanOncallHandoversJobArgs]) error {
		ids, scanErr := oncall.ScanForShiftsNeedingHandover(ctx)
		if scanErr != nil {
			return fmt.Errorf("failed to scan for shifts: %w", scanErr)
		}

		if len(ids) == 0 {
			return nil
		}

		params := make([]river.InsertManyParams, len(ids))
		for i, shiftId := range ids {
			params[i] = river.InsertManyParams{
				Args: ensureShiftHandoverJobArgs{shiftId: shiftId},
				InsertOpts: &river.InsertOpts{
					UniqueOpts: river.UniqueOpts{
						ByArgs: true,
					},
				},
			}
		}

		_, insertErr := s.client.InsertMany(ctx, params)
		if insertErr != nil {
			return fmt.Errorf("could not insert jobs: %w", insertErr)
		}
		return nil
	})

	return river.AddWorkerSafely(s.clientCfg.Workers, worker)
}

type syncProviderDataJobArgs struct {
	Users     bool
	Incidents bool
	Oncall    bool
	Alerts    bool
}

func (syncProviderDataJobArgs) Kind() string {
	return "sync-provider-data"
}

func (s *JobService) registerProviderDataSyncPeriodicJob(
	interval time.Duration,
	users rez.UserService,
	incidents rez.IncidentService,
	oncall rez.OncallService,
	alerts rez.AlertsService,
) error {
	args := &syncProviderDataJobArgs{
		Users:     true,
		Incidents: true,
		Oncall:    true,
		Alerts:    true,
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

	return river.AddWorkerSafely(s.clientCfg.Workers, river.WorkFunc(func(ctx context.Context, j *river.Job[syncProviderDataJobArgs]) error {
		var err error
		if j.Args.Users {
			err = errors.Join(err, users.SyncData(ctx))
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
		return err
	}))
}
