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
		jobs.RegisterWorker(s.jobClient, &checkShiftHandoverJobWorker{svc: s}),
	)
}

// Handover Reminder Job
type checkShiftHandoverJobArgs struct {
	shiftId uuid.UUID
}

func (a checkShiftHandoverJobArgs) Kind() string {
	return "oncall-handover-check-job-args"
}

type checkShiftHandoverJobWorker struct {
	river.WorkerDefaults[checkShiftHandoverJobArgs]
	svc *OncallService
}

func (w *checkShiftHandoverJobWorker) Work(ctx context.Context, job *river.Job[checkShiftHandoverJobArgs]) error {
	return w.svc.checkShiftHandover(ctx, job.Args.shiftId)
}
