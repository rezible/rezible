package db

import (
	"context"
	"fmt"

	"github.com/rezible/rezible/pkg/jobs"
	"github.com/riverqueue/river"
)

func NewProcessProviderEventWorker(service *ProviderEventPipelineService) jobs.WorkerDefinition {
	return jobs.DefineWorkerFunc(service.HandleProcessEventJob)
}

func (s *ProviderEventPipelineService) HandleProcessEventJob(ctx context.Context, args *jobs.ProcessProviderEventArgs) error {
	res := s.processProviderEvent(ctx, args.Event)
	s.telemetry.recordProcessed(ctx, args.Event, res)
	return res.error
}

func NewProjectNormalizedEventWorker(service *ProviderEventPipelineService) jobs.WorkerDefinition {
	return jobs.DefineWorkerFunc(service.HandleEventProjectionJob)
}

func (s *ProviderEventPipelineService) HandleEventProjectionJob(ctx context.Context, args jobs.ProjectNormalizedEvent) error {
	ev, queryErr := s.db.Client(ctx).NormalizedEvent.Get(ctx, args.EventId)
	if queryErr != nil {
		return fmt.Errorf("query event: %w", queryErr)
	}

	projectionErr := s.projectNormalizedEvent(ctx, ev)
	if projectionErr == nil {
		return projectionErr
	}
	if s.db.IsTransientError(projectionErr) {
		return jobs.MarkRetryableError(projectionErr)
	} else if jobs.IsRetryableError(projectionErr) {
		return projectionErr
	}
	s.logger.ErrorContext(ctx, "fatal event projection error",
		"error", projectionErr.Error(),
		"ref", ev.ProviderEventRef,
	)
	return river.JobCancel(projectionErr)
}
