package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/jobs"
)

func NewProcessProviderEventWorker(service *ProviderEventPipelineService) jobs.WorkerDefinition {
	return jobs.DefineWorkerFunc(service.HandleProcessEventJob)
}

func (s *ProviderEventPipelineService) HandleProcessEventJob(ctx context.Context, args *jobs.ProcessProviderEventArgs) error {
	start := time.Now()
	savedCount, processErr := ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (int, error) {
		return s.processAndQueueProjections(ctx, args.Event)
	})
	s.telemetry.recordProcessed(ctx, args.Event, time.Since(start), savedCount, processErr)
	return processErr
}

func (s *ProviderEventPipelineService) processAndQueueProjections(ctx context.Context, ev rez.ProviderEvent) (int, error) {
	ids, ingestErr := s.ingester.ProcessProviderEvent(ctx, ev)
	if ingestErr != nil {
		return 0, ingestErr
	}
	savedCount := len(ids)
	var jobsErr error
	if savedCount > 0 {
		params := make([]river.InsertManyParams, savedCount)
		for i, id := range ids {
			params[i] = river.InsertManyParams{
				Args: jobs.ProjectNormalizedEvent{EventId: id},
			}
		}
		if _, insertErr := s.jobs.InsertMany(ctx, params); insertErr != nil {
			jobsErr = fmt.Errorf("inserting project events: %w", insertErr)
		}
	}
	return savedCount, jobsErr
}

func NewProjectNormalizedEventWorker(service *ProviderEventPipelineService) jobs.WorkerDefinition {
	return jobs.DefineWorkerFunc(service.HandleEventProjectionJob)
}

func (s *ProviderEventPipelineService) HandleEventProjectionJob(ctx context.Context, args jobs.ProjectNormalizedEvent) error {
	projectionErr := s.ingester.ProjectNormalizedEvent(ctx, args.EventId)
	if projectionErr == nil {
		return projectionErr
	}
	if s.db.IsTransientError(projectionErr) {
		return jobs.MarkRetryableError(projectionErr)
	} else if jobs.IsRetryableError(projectionErr) {
		return projectionErr
	}
	slog.ErrorContext(ctx, "fatal event projection error",
		"error", projectionErr.Error(),
		"event_id", args.EventId,
	)
	return river.JobCancel(projectionErr)
}
