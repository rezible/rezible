package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	in "github.com/rezible/rezible/ent/integration"
	iesr "github.com/rezible/rezible/ent/integrationeventsyncrun"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/riverqueue/river"
)

type IntegrationEventsSyncWorker struct {
	river.WorkerDefaults[jobs.SyncIntegrationSourceEvents]

	db       rez.Database
	msgs     rez.MessageService
	intgs    rez.IntegrationService
	registry rez.IntegrationRegistry
	pipeline rez.ProviderEventPipelineService

	logger  *slog.Logger
	timeout time.Duration
}

func NewIntegrationEventsSyncWorker(cfg rez.Config, tel rez.TelemetryService, db rez.Database, msgs rez.MessageService, intgs rez.IntegrationService, reg rez.IntegrationRegistry, pipeline rez.ProviderEventPipelineService) (*IntegrationEventsSyncWorker, error) {
	w := &IntegrationEventsSyncWorker{
		db:       db,
		msgs:     msgs,
		intgs:    intgs,
		registry: reg,
		pipeline: pipeline,
		logger:   tel.NewLogger(rez.NewLoggerOptions{Name: "sync_integration_events_worker"}),
		timeout:  time.Minute * 10,
	}
	return w, nil
}

func (w *IntegrationEventsSyncWorker) Timeout(*river.Job[jobs.SyncIntegrationSourceEvents]) time.Duration {
	return w.timeout
}

func (w *IntegrationEventsSyncWorker) Work(ctx context.Context, job *river.Job[jobs.SyncIntegrationSourceEvents]) error {
	args := job.Args
	if args.IntegrationId == uuid.Nil {
		// TODO: sync all installed?
		return nil
	}

	intg, intgErr := w.intgs.LookupInstallation(ctx, in.ID(args.IntegrationId))
	if intgErr != nil {
		slog.WarnContext(ctx, "failed to get installed integration")
		if ent.IsNotFound(intgErr) {
			return nil
		}
		return fmt.Errorf("get installed integration: %w", intgErr)
	}
	ii, iiErr := w.intgs.AsInstalledIntegration(intg)
	if iiErr != nil {
		return fmt.Errorf("get installed integration: %w", iiErr)
	}

	querier, querierErr := w.registry.GetProviderEventQuerier(ii)
	if querierErr != nil || querier == nil {
		slog.WarnContext(ctx, "failed to get integration event querier", "error", querierErr)
		return nil
	}

	cursors, cursorsErr := w.lookupSourceSyncCursors(ctx, args)
	if cursorsErr != nil {
		slog.WarnContext(ctx, "failed to lookup integration sync cursors", "error", cursorsErr)
		return cursorsErr
	}

	res := w.pipeline.SyncEvents(ctx, querier, cursors)
	if saveResErr := w.saveSyncResult(ctx, args, res); saveResErr != nil {
		slog.ErrorContext(ctx, "failed to save integration event sync run", "error", saveResErr)
	}
	return nil
}

func (w *IntegrationEventsSyncWorker) lookupSourceSyncCursors(ctx context.Context, args jobs.SyncIntegrationSourceEvents) (rez.ProviderEventSourceCursors, error) {
	sourceCursors := rez.ProviderEventSourceCursors{}
	for _, src := range args.Sources {
		// TODO: look up cursors from last sync
		sourceCursors[src] = ""
	}

	return sourceCursors, nil
}

func (w *IntegrationEventsSyncWorker) saveSyncResult(ctx context.Context, args jobs.SyncIntegrationSourceEvents, res rez.ProviderEventSyncResult) error {
	// TODO: we should store this properly, but will require refactoring the integration sync result struct
	startedAt := time.Now()
	for _, dur := range res.SourceSyncDurations {
		startedAt = startedAt.Add(-dur)
	}
	saveRun := w.db.Client(ctx).IntegrationEventSyncRun.Create().
		SetSyncReason(args.SyncReason).
		SetIntegrationID(args.IntegrationId).
		SetStartedAt(startedAt).
		SetFinishedAt(time.Now().UTC()).
		SetStatus(iesr.StatusSuccess).
		SetProviderEventSourceCursors(res.SourceCursorsAfter).
		SetEventsPulled(res.EventsPulled).
		SetEventsIngested(res.EventsIngested).
		SetDuplicates(res.NumDuplicates)
	if len(res.SyncErrors) > 0 {
		saveRun.SetStatus(iesr.StatusFailed)
		saveRun.SetFailureMessage(errors.Join(res.SyncErrors...).Error())
	}
	return saveRun.Exec(ctx)
}
