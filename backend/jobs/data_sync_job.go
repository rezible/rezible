package jobs

import (
	"context"
	"errors"
	"github.com/riverqueue/river"
	rez "github.com/twohundreds/rezible"
	"time"
)

func RegisterProviderDataSyncJob(
	client *BackgroundJobClient,
	users rez.UserService,
	incidents rez.IncidentService,
	oncall rez.OncallService,
	alerts rez.AlertsService,
) error {
	worker := &dataSyncJobWorker{
		users:     users,
		incidents: incidents,
		oncall:    oncall,
		alerts:    alerts,
	}

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

	return RegisterWorker(client, worker)
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
