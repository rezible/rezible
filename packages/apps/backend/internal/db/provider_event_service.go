package db

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/rezible/rezible/integrations/eventprojections"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	neps "github.com/rezible/rezible/ent/normalizedeventprojectionstatus"
	pesc "github.com/rezible/rezible/ent/providereventsynccursor"
	pesr "github.com/rezible/rezible/ent/providereventsyncrun"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/jobs"
	"github.com/rezible/rezible/telemetry"
)

const providerEventSyncBatchSize = 100

type ProviderEventService struct {
	logger *slog.Logger
	db     *ent.Client

	jobService   rez.JobsService
	integrations rez.IntegrationsService
	metrics      *providerEventMetrics

	processorsMu sync.RWMutex
	processors   map[string]rez.ProviderEventProcessor
}

func NewProviderEventService(ctx context.Context, svcs *rez.Services) *ProviderEventService {
	pe := &ProviderEventService{
		logger:       telemetry.NewLogger(ctx, telemetry.WithLogPackage("provider_events")),
		db:           svcs.Database.Client(),
		jobService:   svcs.Jobs,
		integrations: svcs.Integrations,
		metrics:      newProviderEventMetrics(),
		processorsMu: sync.RWMutex{},
		processors:   make(map[string]rez.ProviderEventProcessor),
	}
	jobs.RegisterWorkerFunc(pe.HandleProviderEventSyncJob)
	jobs.RegisterWorkerFunc(pe.ProcessProviderEvent)
	jobs.RegisterWorkerFunc(pe.ProjectNormalizedEvent)
	return pe
}

func (s *ProviderEventService) RegisterEventProcessor(provider string, proc rez.ProviderEventProcessor) {
	s.processorsMu.Lock()
	defer s.processorsMu.Unlock()
	s.processors[provider] = proc
}

func (s *ProviderEventService) lookupEventProcessor(provider string) (rez.ProviderEventProcessor, bool) {
	s.processorsMu.RLock()
	defer s.processorsMu.RUnlock()
	proc, ok := s.processors[provider]
	return proc, ok
}

func (s *ProviderEventService) Ingest(ctx context.Context, ev rez.ProviderEvent) (*rez.ProviderEventIngestResult, error) {
	res, ingestErr := s.ingest(ctx, ev)
	s.metrics.recordIngested(ctx, ev, res, ingestErr)
	return res, ingestErr
}

func (s *ProviderEventService) ingest(ctx context.Context, ev rez.ProviderEvent) (*rez.ProviderEventIngestResult, error) {
	processEvent, validationErr := s.makeProcessEventArgs(ev)
	if validationErr != nil {
		return nil, fmt.Errorf("invalid event: %w", validationErr)
	}

	insertOpts := &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{ByArgs: true},
	}
	insertRes, insertErr := s.jobService.Insert(ctx, *processEvent, insertOpts)
	if insertErr != nil {
		return nil, fmt.Errorf("could not insert provider event job: %w", insertErr)
	}

	res := &rez.ProviderEventIngestResult{
		Duplicate: insertRes != nil && insertRes.UniqueSkippedAsDuplicate,
	}
	if res.Duplicate {
		s.logger.Debug("skipped duplicate provider event",
			"provider", ev.Provider,
			"source", ev.ProviderSource,
		)
	}
	return res, nil
}

type processProviderEventArgs struct {
	Event rez.ProviderEvent
}

func (processProviderEventArgs) Kind() string {
	return "process-provider-event"
}

func (s *ProviderEventService) makeProcessEventArgs(ev rez.ProviderEvent) (*processProviderEventArgs, error) {
	args := processProviderEventArgs{Event: ev}
	if ev.ReceivedAt.IsZero() {
		args.Event.ReceivedAt = time.Now().UTC()
	}
	if ev.Provider == "" {
		return nil, fmt.Errorf("provider event provider is required")
	}
	if ev.ProviderSource == "" {
		return nil, fmt.Errorf("provider event provider_source is required")
	}
	if ev.SubjectRef == "" {
		return nil, fmt.Errorf("provider event subject_ref is required")
	}
	if len(ev.Payload) == 0 {
		return nil, fmt.Errorf("provider event payload is required")
	}
	if ev.ProviderEventRef == "" {
		return nil, fmt.Errorf("provider event provider_delivery_ref is required")
	}
	return &args, nil
}

func (s *ProviderEventService) ProcessProviderEvent(ctx context.Context, args processProviderEventArgs) error {
	if _, tenantOk := execution.GetContext(ctx).TenantID(); !tenantOk {
		return fmt.Errorf("tenant not found in context")
	}
	res, err := s.processProviderEvent(ctx, args)
	s.metrics.recordProcessed(ctx, args.Event, res, err)
	return err
}

type processProviderEventResult struct {
	processTime    time.Duration
	processSuccess bool
	normalizeCount int

	projectionSuccess bool
	projectionTime    time.Duration
}

func (s *ProviderEventService) processProviderEvent(ctx context.Context, args processProviderEventArgs) (*processProviderEventResult, error) {
	processStart := time.Now()
	proc, procExists := s.lookupEventProcessor(args.Event.Provider)
	if !procExists {
		return nil, fmt.Errorf("no provider event processor registered for provider %s", args.Event.Provider)
	}

	res := &processProviderEventResult{}

	normalizedEvents, procErr := proc.Process(ctx, args.Event)
	res.processTime = time.Since(processStart)
	res.normalizeCount = len(normalizedEvents)
	if procErr != nil {
		return res, fmt.Errorf("processing provider event: %w", procErr)
	}
	res.processSuccess = true

	if res.normalizeCount > 0 {
		projectionStart := time.Now()
		projErr := s.saveNormalizedEvents(ctx, normalizedEvents)
		res.projectionTime = time.Since(projectionStart)
		if projErr != nil {
			return res, fmt.Errorf("saving normalized events: %w", projErr)
		}
		res.projectionSuccess = true
	}

	return res, nil
}

func (s *ProviderEventService) saveNormalizedEvents(ctx context.Context, normalizedEvents ent.NormalizedEvents) error {
	conflictCols := sql.ConflictColumns(
		ne.FieldProvider,
		ne.FieldProviderSource,
		ne.FieldProviderEventRef,
		ne.FieldKind,
		ne.FieldSubjectRef,
	)
	return ent.WithTx(ctx, s.db, func(tx *ent.Tx) error {
		createBulk := tx.NormalizedEvent.MapCreateBulk(normalizedEvents, func(c *ent.NormalizedEventCreate, i int) {
			ev := normalizedEvents[i]
			c.SetProvider(ev.Provider).
				SetProviderSource(ev.ProviderSource).
				SetProviderEventRef(ev.ProviderEventRef).
				SetKind(ev.Kind).
				SetSubjectKind(ev.SubjectKind).
				SetSubjectRef(ev.SubjectRef).
				SetOccurredAt(ev.OccurredAt).
				SetReceivedAt(ev.ReceivedAt).
				SetAttributes(ev.Attributes).
				OnConflict(conflictCols).
				UpdateNewValues()
		})

		evs, upsertErr := createBulk.Save(ctx)
		if upsertErr != nil {
			return fmt.Errorf("upsert normalized events: %w", upsertErr)
		}

		params := make([]river.InsertManyParams, len(evs))
		for i, ev := range evs {
			params[i] = river.InsertManyParams{
				Args:       jobs.ProjectNormalizedEvent{EventId: ev.ID},
				InsertOpts: &river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}},
			}
		}
		res, jobErr := s.jobService.InsertManyTx(ctx, tx, params)
		if jobErr != nil {
			return fmt.Errorf("inserting project events: %w", jobErr)
		}
		slog.DebugContext(ctx, "requested projection for normalized events",
			"count", len(res))

		return nil
	})
}

func (s *ProviderEventService) HandleProviderEventSyncJob(ctx context.Context, args jobs.ProviderEventSyncJob) error {
	if len(args.ProviderSources) == 0 {
		slog.WarnContext(ctx, "TODO: support syncing all providers")
	}

	for provider, sources := range args.ProviderSources {
		queriers, discoverErr := s.integrations.GetProviderEventQueriers(ctx, provider)
		if discoverErr != nil {
			return fmt.Errorf("discover provider event queriers: %w", discoverErr)
		}

		matched := make([]rez.ProviderEventQuerier, 0, len(queriers))
		for _, querier := range queriers {
			if provider != "" && querier.Provider() != provider {
				continue
			}
			if len(sources) > 0 && !slices.Contains(sources, querier.ProviderSource()) {
				continue
			}
			matched = append(matched, querier)
		}

		for _, querier := range matched {
			syncOpts := rez.ProviderEventSyncOptions{
				CursorAfter: args.CursorAfter,
				SyncReason:  args.SyncReason,
			}
			if syncErr := s.SyncEvents(ctx, querier, syncOpts); syncErr != nil {
				return fmt.Errorf("sync: %w", syncErr)
			}
		}
	}
	return nil
}

type providerEventSyncStats struct {
	eventsPulled   int
	eventsIngested int
	duplicates     int
	latestCursor   *string
}

func (s *ProviderEventService) SyncEvents(ctx context.Context, querier rez.ProviderEventQuerier, opts rez.ProviderEventSyncOptions) error {
	var cursorAfter string
	if opts.CursorAfter != nil {
		cursorAfter = *opts.CursorAfter
	}

	startedAt := time.Now()
	stats, syncErr := s.syncEvents(ctx, querier, cursorAfter)

	createRun := s.db.ProviderEventSyncRun.Create().
		SetProvider(querier.Provider()).
		SetProviderSource(querier.ProviderSource()).
		SetSyncReason(opts.SyncReason).
		SetStartedAt(startedAt).
		SetFinishedAt(time.Now().UTC()).
		SetEventsPulled(stats.eventsPulled).
		SetEventsIngested(stats.eventsIngested).
		SetDuplicates(stats.duplicates).
		SetStatus(pesr.StatusSuccess)
	if syncErr != nil {
		createRun.SetStatus(pesr.StatusFailed)
		errMsg := strings.TrimSpace(syncErr.Error())
		if len(errMsg) > 100 {
			errMsg = errMsg[:100] + "..."
		}
		createRun.SetFailureMessage(errMsg)
	}
	if _, updateErr := createRun.Save(ctx); updateErr != nil {
		return fmt.Errorf("save provider event sync run status: %w", updateErr)
	}

	return syncErr
}

func (s *ProviderEventService) syncEvents(ctx context.Context, querier rez.ProviderEventQuerier, cursorAfter string) (providerEventSyncStats, error) {
	var stats providerEventSyncStats
	if cursorAfter == "" {
		cursor, cursorErr := s.loadSyncCursor(ctx, querier)
		if cursorErr != nil {
			return stats, fmt.Errorf("load cursor: %w", cursorErr)
		} else if cursor != nil {
			cursorAfter = cursor.Cursor
		}
	}

	type syncBatchItem struct {
		args        processProviderEventArgs
		cursorAfter *string
	}

	batch := make([]syncBatchItem, 0, providerEventSyncBatchSize)

	flushBatch := func() error {
		if len(batch) == 0 {
			return nil
		}
		params := make([]river.InsertManyParams, len(batch))
		for i, item := range batch {
			params[i] = river.InsertManyParams{
				Args:       item.args,
				InsertOpts: &river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}},
			}
		}
		results, insertErr := s.jobService.InsertMany(ctx, params)
		if insertErr != nil {
			return fmt.Errorf("could not insert provider event jobs: %w", insertErr)
		}
		for i, result := range results {
			if result != nil && result.UniqueSkippedAsDuplicate {
				stats.duplicates++
			} else {
				stats.eventsIngested++
			}
			if i < len(batch) {
				if cursor := *batch[i].cursorAfter; cursor != "" {
					stats.latestCursor = &cursor
				}
			}
		}
		if stats.latestCursor != nil {
			if cursorErr := s.saveSyncCursor(ctx, querier, *stats.latestCursor); cursorErr != nil {
				return cursorErr
			}
		}
		return nil
	}

	req := rez.ProviderEventQueryRequest{CursorAfter: cursorAfter}
	for result, pullErr := range querier.PullEvents(ctx, req) {
		if pullErr != nil {
			return stats, fmt.Errorf("pull provider events: %w", pullErr)
		}
		stats.eventsPulled++
		args, validErr := s.makeProcessEventArgs(result.Event)
		if validErr != nil {
			return stats, fmt.Errorf("invalid event args: %w", validErr)
		}
		batch = append(batch, syncBatchItem{args: *args, cursorAfter: result.CursorAfter})
		if len(batch) >= providerEventSyncBatchSize {
			if flushErr := flushBatch(); flushErr != nil {
				return stats, flushErr
			}
			batch = make([]syncBatchItem, 0, providerEventSyncBatchSize)
		}
		if result.CursorAfter == nil {
			break
		}
	}
	return stats, flushBatch()
}

func (s *ProviderEventService) loadSyncCursor(ctx context.Context, q rez.ProviderEventQuerier) (*ent.ProviderEventSyncCursor, error) {
	query := s.db.ProviderEventSyncCursor.Query().
		Where(pesc.Provider(q.Provider()), pesc.ProviderSource(q.ProviderSource()))
	cursor, queryErr := query.Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return nil, fmt.Errorf("query provider event sync cursor: %w", queryErr)
	}
	return cursor, nil
}

func (s *ProviderEventService) saveSyncCursor(ctx context.Context, querier rez.ProviderEventQuerier, cursor string) error {
	create := s.db.ProviderEventSyncCursor.Create().
		SetProvider(querier.Provider()).
		SetProviderSource(querier.ProviderSource()).
		SetCursor(cursor).
		SetLastSyncedAt(time.Now().UTC())

	create.OnConflict(sql.ConflictColumns(
		pesc.FieldProvider,
		pesc.FieldProviderSource,
	)).
		Update(func(u *ent.ProviderEventSyncCursorUpsert) {
			u.UpdateCursor()
			u.UpdateLastSyncedAt()
			u.UpdateUpdatedAt()
		})

	return create.Exec(ctx)
}

func (s *ProviderEventService) ProjectNormalizedEvent(ctx context.Context, args jobs.ProjectNormalizedEvent) error {
	ev, queryErr := s.db.NormalizedEvent.Get(ctx, args.EventId)
	if queryErr != nil {
		return fmt.Errorf("query normalized event: %w", queryErr)
	}

	var failed []error
	for name, handlerFn := range eventprojections.GetHandlers() {
		queryStatus := s.db.NormalizedEventProjectionStatus.Query().
			Where(neps.NormalizedEventID(ev.ID), neps.HandlerName(name))
		status, statusErr := queryStatus.Only(ctx)
		if statusErr != nil && !ent.IsNotFound(statusErr) {
			failed = append(failed, fmt.Errorf("query projection status: %w", statusErr))
			continue
		}

		if status != nil {
			if status.Status == neps.StatusSucceeded || status.Status == neps.StatusPending {
				continue
			}
		} else {
			status, statusErr = s.db.NormalizedEventProjectionStatus.Create().
				SetStatus(neps.StatusPending).
				SetNormalizedEventID(args.EventId).
				SetHandlerName(name).
				SetLastAttemptedAt(time.Now().UTC()).
				Save(ctx)
			if statusErr != nil {
				failed = append(failed, fmt.Errorf("mark projection pending: %w", statusErr))
				continue
			}
		}

		projectErr := ent.WithTx(ctx, s.db, func(tx *ent.Tx) error {
			return handlerFn(ctx, tx.Client(), ev)
		})

		update := status.Update()
		if projectErr == nil {
			update.SetStatus(neps.StatusSucceeded).
				SetSucceededAt(time.Now().UTC()).
				ClearFailedAt().
				ClearLastError()
		} else {
			update.SetStatus(neps.StatusFailed).
				SetFailedAt(time.Now().UTC()).
				SetLastError(projectErr.Error())
		}
		if statusErr = update.Exec(ctx); statusErr != nil {
			failed = append(failed, fmt.Errorf("update projection status: %w", statusErr))
			continue
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("project normalized event: %v", failed)
	}
	return nil
}

type providerEventMetrics struct {
	ingested          telemetry.Int64Counter
	processed         telemetry.Int64Counter
	processSeconds    telemetry.Float64Histogram
	projectionSeconds telemetry.Float64Histogram
	normalizedEvents  telemetry.Int64Counter
}

func newProviderEventMetrics() *providerEventMetrics {
	meter := telemetry.DefaultMeter()
	return &providerEventMetrics{
		ingested:          telemetry.Int64CounterInstrument(meter, "rezible.backend.provider_events.ingested", "Provider events ingested"),
		processed:         telemetry.Int64CounterInstrument(meter, "rezible.backend.provider_events.processed", "Provider events processed"),
		processSeconds:    telemetry.Float64HistogramInstrument(meter, "rezible.backend.provider_events.normalize_duration", "Provider event normalization processing duration", "s"),
		normalizedEvents:  telemetry.Int64CounterInstrument(meter, "rezible.backend.provider_events.normalized_events", "Normalized provider events saved"),
		projectionSeconds: telemetry.Float64HistogramInstrument(meter, "rezible.backend.provider_events.projection_duration", "Normalized event projection duration", "s"),
	}
}

func (m *providerEventMetrics) recordIngested(ctx context.Context, ev rez.ProviderEvent, res *rez.ProviderEventIngestResult, err error) {
	if m != nil {
		m.ingested.Add(ctx, 1, telemetry.WithMetricAttributes(
			telemetry.StringAttr("provider", telemetry.NormalizeLabel(ev.Provider)),
			telemetry.StringAttr("provider_source", telemetry.NormalizeLabel(ev.ProviderSource)),
			telemetry.ResultAttr(err),
			telemetry.BoolAttr("duplicate", res != nil && res.Duplicate),
		))
	}
}

func (m *providerEventMetrics) recordProcessed(ctx context.Context, ev rez.ProviderEvent, res *processProviderEventResult, err error) {
	if m != nil {
		processSuccess := res != nil && res.processSuccess
		projectionSuccess := res != nil && res.projectionSuccess
		attrs := []telemetry.KeyValue{
			telemetry.StringAttr("provider", telemetry.NormalizeLabel(ev.Provider)),
			telemetry.StringAttr("provider_source", telemetry.NormalizeLabel(ev.ProviderSource)),
			telemetry.ResultAttr(err),
			telemetry.BoolAttr("process_success", processSuccess),
			telemetry.BoolAttr("projection_success", projectionSuccess),
		}
		m.processed.Add(ctx, 1, telemetry.WithMetricAttributes(attrs...))
		if res != nil {
			m.processSeconds.Record(ctx, res.processTime.Seconds(), telemetry.WithMetricAttributes(attrs...))
			if res.normalizeCount > 0 {
				m.normalizedEvents.Add(ctx, int64(res.normalizeCount), telemetry.WithMetricAttributes(attrs...))
				m.projectionSeconds.Record(ctx, res.projectionTime.Seconds(), telemetry.WithMetricAttributes(attrs...))
			}
		}
	}
}
