package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/rezible/rezible/jobs"
	"github.com/riverqueue/river"
)

func (s *IncidentService) RegisterJobs() error {
	return errors.Join(
		jobs.RegisterWorker(s.jobClient, &generateIncidentDebriefResponseWorker{svc: s}),
		jobs.RegisterWorker(s.jobClient, &sendIncidentDebriefRequestsWorker{svc: s}),
	)
}

type sendIncidentDebriefRequestsJobArgs struct {
	IncidentId uuid.UUID
}

func (sendIncidentDebriefRequestsJobArgs) Kind() string {
	return "send-incident-debrief-requests"
}

type sendIncidentDebriefRequestsWorker struct {
	river.WorkerDefaults[sendIncidentDebriefRequestsJobArgs]
	svc *IncidentService
}

func (w *sendIncidentDebriefRequestsWorker) Work(ctx context.Context, job *river.Job[sendIncidentDebriefRequestsJobArgs]) error {
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
	svc *IncidentService
}

func (w *generateIncidentDebriefResponseWorker) Work(ctx context.Context, job *river.Job[generateIncidentDebriefResponseJobArgs]) error {
	return w.svc.generateDebriefMessage(ctx, job.Args.DebriefId)
}
