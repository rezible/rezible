package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/rezible/rezible/jobs"
	"github.com/riverqueue/river"
)

func (s *DebriefService) RegisterJobWorkers() error {
	return errors.Join(
		jobs.RegisterWorker(s.jobClient, &sendIncidentDebriefRequestsWorker{svc: s}),
		jobs.RegisterWorker(s.jobClient, &generateIncidentDebriefResponseWorker{svc: s}),
	)
}

type sendIncidentDebriefRequestsWorker struct {
	river.WorkerDefaults[jobs.SendIncidentDebriefRequestsJobArgs]
	svc *DebriefService
}

func (w *sendIncidentDebriefRequestsWorker) Work(ctx context.Context, job *river.Job[jobs.SendIncidentDebriefRequestsJobArgs]) error {
	return w.svc.sendDebriefRequests(ctx, job.Args.IncidentId)
}

type generateIncidentDebriefResponseJobArgs struct {
	DebriefId uuid.UUID
}

func (generateIncidentDebriefResponseJobArgs) Kind() string {
	return "generate-incident-debrief-response"
}

type generateIncidentDebriefResponseWorker struct {
	river.WorkerDefaults[generateIncidentDebriefResponseJobArgs]
	svc *DebriefService
}

func (w *generateIncidentDebriefResponseWorker) Work(ctx context.Context, job *river.Job[generateIncidentDebriefResponseJobArgs]) error {
	return w.svc.generateDebriefResponse(ctx, job.Args.DebriefId)
}
