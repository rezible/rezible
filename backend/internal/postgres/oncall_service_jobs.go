package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/twohundreds/rezible/jobs"
)

func (s *OncallService) RegisterJobs() error {
	//s.jobClient.AddPeriodicJob(jobs.PeriodicJob{
	//	Schedule: jobs.PeriodicInterval(time.Hour),
	//	Constructor: func() (jobs.JobArgs, *jobs.InsertOpts) {
	//		return &oncallDataSyncJobArgs{}, nil
	//	},
	//	Options: &jobs.PeriodicJobOpts{
	//		RunOnStart: true,
	//	},
	//})

	return errors.Join(
		//jobs.RegisterWorker(s.jobClient, &oncallDataSyncJobWorker{ds: s.dataSyncer}),
		jobs.RegisterWorker(s.jobClient, &oncallHandoverReminderJobWorker{svc: s}),
	)
}

// Data Provider Sync Job
type oncallDataSyncJobArgs struct{}

func (a oncallDataSyncJobArgs) Kind() string {
	return "sync-oncall-data"
}

type oncallDataSyncJobWorker struct {
	river.WorkerDefaults[oncallDataSyncJobArgs]
	ds *oncallDataSyncer
}

func (w *oncallDataSyncJobWorker) Work(ctx context.Context, job *river.Job[oncallDataSyncJobArgs]) error {
	return w.ds.syncProviderData(ctx)
}

// Handover Reminder Job
type oncallHandoverReminderJobArgs struct {
	shiftId uuid.UUID
}

func (a oncallHandoverReminderJobArgs) Kind() string {
	return "oncall-handover-reminder-job-args"
}

type oncallHandoverReminderJobWorker struct {
	river.WorkerDefaults[oncallHandoverReminderJobArgs]
	svc *OncallService
}

func (w *oncallHandoverReminderJobWorker) Work(ctx context.Context, job *river.Job[oncallHandoverReminderJobArgs]) error {
	return w.svc.sendShiftHandoverReminder(ctx, job.Args.shiftId)
}
