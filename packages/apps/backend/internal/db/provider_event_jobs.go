package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/errs"
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
	result, ingestErr := s.ingester.ProcessProviderEvent(ctx, ev)
	if ingestErr != nil {
		return 0, ingestErr
	}
	savedCount := len(result.NewEventIDs)
	var jobsErr error
	if savedCount+len(result.Reobservations) > 0 {
		params := make([]river.InsertManyParams, savedCount)
		for i, id := range result.NewEventIDs {
			params[i] = river.InsertManyParams{
				Args: jobs.ProjectNormalizedEvent{EventId: id},
			}
		}
		for _, repeat := range result.Reobservations {
			params = append(params, river.InsertManyParams{Args: jobs.ReobserveNormalizedEvent{EventId: repeat.EventID, At: repeat.At}})
		}
		if _, insertErr := s.jobs.InsertMany(ctx, params); insertErr != nil {
			jobsErr = fmt.Errorf("inserting project events: %w", insertErr)
		}
	}
	return savedCount, jobsErr
}

// The fifth projection attempt records missing endpoints instead of retrying them. It must stay below
// jobs.ProjectNormalizedEventMaxAttempts.
const missingEndpointSkipAttempt = 5

func NewProjectNormalizedEventWorker(service *ProviderEventPipelineService) jobs.WorkerDefinition {
	return jobs.DefineWorker(river.WorkFunc(service.HandleEventProjectionJob))
}

func (s *ProviderEventPipelineService) HandleEventProjectionJob(ctx context.Context, job *river.Job[jobs.ProjectNormalizedEvent]) error {
	return s.classifyEventJobError(ctx, job.Args.EventId, "projection", s.ingester.ProjectNormalizedEvent(ctx, job.Args.EventId, job.Attempt < missingEndpointSkipAttempt))
}

func (s *ProviderEventPipelineService) classifyEventJobError(ctx context.Context, eventID uuid.UUID, operation string, jobErr error) error {
	if jobErr == nil {
		return nil
	}
	// A merge that moved a resolved subject's aliases while the job waited for its locks leaves either worker to
	// resolve it again.
	if s.db.IsTransientError(jobErr) || errors.Is(jobErr, errs.ErrKnowledgeSubjectMerged) {
		return jobs.MarkRetryableError(jobErr)
	}
	// A unique violation means a concurrent job created the row this one looked up and did not find. The
	// retry finds it.
	if _, isUnique := s.db.IsConstraintError(jobErr); isUnique {
		return jobs.MarkRetryableError(jobErr)
	}
	// The worker was stopped, or the job timed out, before the work finished.
	if errors.Is(jobErr, context.Canceled) || errors.Is(jobErr, context.DeadlineExceeded) {
		return jobs.MarkRetryableError(jobErr)
	}
	if jobs.IsRetryableError(jobErr) {
		return jobErr
	}
	slog.ErrorContext(ctx, "fatal event job error",
		"operation", operation,
		"error", jobErr.Error(),
		"event_id", eventID,
	)
	return river.JobCancel(jobErr)
}

func NewReobserveNormalizedEventWorker(service *ProviderEventPipelineService) jobs.WorkerDefinition {
	return jobs.DefineWorkerFunc(service.HandleEventReobservationJob)
}

func (s *ProviderEventPipelineService) HandleEventReobservationJob(ctx context.Context, args jobs.ReobserveNormalizedEvent) error {
	return s.classifyEventJobError(ctx, args.EventId, "re-observation", s.ingester.ReobserveNormalizedEvent(ctx, args.EventId, args.At))
}
