package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/rezible/rezible/jobs"
	"github.com/riverqueue/river"
)

func (s *OncallService) RegisterJobs() error {
	return errors.Join(
		jobs.RegisterWorker(s.jobClient, &oncallHandoverReminderJobWorker{svc: s}),
	)
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
