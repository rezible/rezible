package jobs

import (
	"context"
	"errors"
	"fmt"
	rez "github.com/rezible/rezible"
	"github.com/riverqueue/river"
	"time"
)

func RegisterPeriodicJobs(
	client *BackgroundJobClient,
	users rez.UserService,
	incidents rez.IncidentService,
	oncall rez.OncallService,
	alerts rez.AlertsService,
) error {
	if syncJobErr := registerProviderDataSyncJob(client, users, incidents, oncall, alerts); syncJobErr != nil {
		return fmt.Errorf("register provider data sync: %w", syncJobErr)
	}

	if handoversErr := registerCheckOncallHandoversJob(client, oncall); handoversErr != nil {
		return fmt.Errorf("register check handovers: %w", handoversErr)
	}

	return nil
}

type checkOncallHandoversJobArgs struct{}

func (checkOncallHandoversJobArgs) Kind() string {
	return "send-oncall-handovers"
}

type checkOncallHandoversJobWorker struct {
	river.WorkerDefaults[checkOncallHandoversJobArgs]
	svc rez.OncallService
}

func (w *checkOncallHandoversJobWorker) Work(ctx context.Context, job *river.Job[checkOncallHandoversJobArgs]) error {
	return w.svc.CheckOncallHandovers(ctx)
}

func registerCheckOncallHandoversJob(client *BackgroundJobClient, oncall rez.OncallService) error {
	client.AddPeriodicJob(PeriodicJob{
		Schedule: PeriodicInterval(time.Hour),
		Constructor: func() (JobArgs, *InsertOpts) {
			return &checkOncallHandoversJobArgs{}, nil
		},
		Options: &PeriodicJobOpts{
			RunOnStart: true,
		},
	})

	return RegisterWorker(client, &checkOncallHandoversJobWorker{svc: oncall})
}

func registerProviderDataSyncJob(
	client *BackgroundJobClient,
	users rez.UserService,
	incidents rez.IncidentService,
	oncall rez.OncallService,
	alerts rez.AlertsService,
) error {
	client.AddPeriodicJob(PeriodicJob{
		Schedule: PeriodicInterval(time.Hour),
		Constructor: func() (JobArgs, *InsertOpts) {
			return &SyncProviderDataJobArgs{
				Users:     true,
				Incidents: true,
				Oncall:    true,
				Alerts:    true,
			}, nil
		},
		Options: &PeriodicJobOpts{
			RunOnStart: true,
		},
	})

	return RegisterWorker(client, &dataSyncJobWorker{
		users:     users,
		incidents: incidents,
		oncall:    oncall,
		alerts:    alerts,
	})
}

type SyncProviderDataJobArgs struct {
	Users     bool
	Incidents bool
	Oncall    bool
	Alerts    bool
}

func (SyncProviderDataJobArgs) Kind() string {
	return "sync-provider-data"
}

type dataSyncJobWorker struct {
	river.WorkerDefaults[SyncProviderDataJobArgs]

	users     rez.UserService
	incidents rez.IncidentService
	oncall    rez.OncallService
	alerts    rez.AlertsService
}

func (w *dataSyncJobWorker) Work(ctx context.Context, job *river.Job[SyncProviderDataJobArgs]) error {
	args := job.Args

	var err error
	if args.Users {
		err = errors.Join(err, w.users.SyncData(ctx))
	}
	if args.Oncall {
		err = errors.Join(err, w.oncall.SyncData(ctx))
	}
	if args.Incidents {
		err = errors.Join(err, w.incidents.SyncData(ctx))
	}
	if args.Alerts {
		err = errors.Join(err, w.alerts.SyncData(ctx))
	}
	return err
}
