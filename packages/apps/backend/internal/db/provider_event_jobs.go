package db

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/riverqueue/river"
)

func (s *ProviderEventPipelineService) RegisterJobs(registry *jobs.Registry) error {
	if registerErr := registry.AddWorkerFunc(s.HandleProcessEventJob); registerErr != nil {
		return fmt.Errorf("process provider events: %w", registerErr)
	}
	if registerErr := registry.AddWorkerFunc(s.HandleEventProjectionJob); registerErr != nil {
		return fmt.Errorf("project normalized events: %w", registerErr)
	}
	return nil
}

type processProviderEventArgs struct {
	Event rez.ProviderEvent
}

func (processProviderEventArgs) Kind() string {
	return "process-provider-event"
}

func (processProviderEventArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}}
}

func (s *ProviderEventPipelineService) HandleProcessEventJob(ctx context.Context, args processProviderEventArgs) error {
	res := s.processProviderEvent(ctx, args.Event)
	s.telemetry.recordProcessed(ctx, args.Event, res)
	return res.error
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
		return projections.Retryable(projectionErr)
	}
	if projections.IsRetryable(projectionErr) {
		return projectionErr
	}
	s.logger.ErrorContext(ctx, "fatal event projection error",
		"error", projectionErr.Error(),
		"ref", ev.ProviderEventRef,
	)
	return river.JobCancel(projectionErr)
}
