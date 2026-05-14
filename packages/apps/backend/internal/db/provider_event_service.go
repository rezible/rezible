package db

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/integrations/eventprojections"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	neps "github.com/rezible/rezible/ent/normalizedeventprojectionstatus"
	pesc "github.com/rezible/rezible/ent/providereventsynccursor"
	pesr "github.com/rezible/rezible/ent/providereventsyncrun"
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
	jobs.RegisterWorkerFunc(pe.HandleProcessEventJob)
	jobs.RegisterWorkerFunc(pe.HandleEventProjectionJob)
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
	processEvent := processProviderEventArgs{Event: ev}
	if validationErr := processEvent.validate(); validationErr != nil {
		return nil, fmt.Errorf("invalid event: %w", validationErr)
	}

	insertOpts := &river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}}
	insertRes, insertErr := s.jobService.Insert(ctx, processEvent, insertOpts)
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

func (s *ProviderEventService) HandleProviderEventSyncJob(ctx context.Context, args jobs.ProviderEventSyncJob) error {
	if len(args.ProviderSources) == 0 {
		slog.WarnContext(ctx, "TODO: support syncing all providers")
		return nil
	}

	var matchedQueriers []rez.ProviderEventQuerier
	for provider, sources := range args.ProviderSources {
		queriers, discoverErr := s.integrations.GetProviderEventQueriers(ctx, provider)
		if discoverErr != nil {
			return fmt.Errorf("discover provider event queriers: %w", discoverErr)
		}

		sourceMap := mapset.NewSet(sources...)
		for _, querier := range queriers {
			providerMatches := querier.Provider() == provider
			sourceMatches := sourceMap.Contains(querier.ProviderSource()) || len(sources) == 0
			if providerMatches && sourceMatches {
				matchedQueriers = append(matchedQueriers, querier)
			}
		}
	}

	syncOpts := rez.ProviderEventSyncOptions{
		CursorAfter: args.CursorAfter,
		SyncReason:  args.SyncReason,
	}
	for _, querier := range matchedQueriers {
		if syncErr := s.SyncEvents(ctx, querier, syncOpts); syncErr != nil {
			return fmt.Errorf("sync: %w", syncErr)
		}
	}

	return nil
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
		SetEventsPulled(stats.pulled).
		SetEventsIngested(stats.ingested).
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

func (s *ProviderEventService) HandleProcessEventJob(ctx context.Context, args processProviderEventArgs) error {
	res, err := s.processProviderEvent(ctx, args.Event)
	s.metrics.recordProcessed(ctx, args.Event, res, err)
	return err
}

func (s *ProviderEventService) HandleEventProjectionJob(ctx context.Context, args jobs.ProjectNormalizedEvent) error {
	ev, queryErr := s.db.NormalizedEvent.Get(ctx, args.EventId)
	if queryErr != nil {
		return fmt.Errorf("query normalized event: %w", queryErr)
	}
	return s.projectNormalizedEvent(ctx, ev)
}

type providerEventSyncStats struct {
	pulled       int
	ingested     int
	duplicates   int
	latestCursor *string
}

func (s *ProviderEventService) syncEvents(ctx context.Context, querier rez.ProviderEventQuerier, cursorAfter string) (providerEventSyncStats, error) {
	var stats providerEventSyncStats

	if cursorAfter == "" {
		query := s.db.ProviderEventSyncCursor.Query().
			Where(pesc.Provider(querier.Provider()), pesc.ProviderSource(querier.ProviderSource()))
		cursorRes, queryErr := query.Only(ctx)
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			return stats, fmt.Errorf("query provider event sync cursor: %w", queryErr)
		} else if cursorRes != nil {
			cursorAfter = cursorRes.Cursor
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
		slog.DebugContext(ctx, "flushing batch", "len", len(batch))
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
				stats.ingested++
			}
			if i < len(batch) {
				fmt.Printf("sync cursor %d %d\n", i, len(batch))
				if cursor := *batch[i].cursorAfter; cursor != "" {
					stats.latestCursor = &cursor
				}
			}
		}
		if stats.latestCursor != nil {
			saveCursor := s.db.ProviderEventSyncCursor.Create().
				SetProvider(querier.Provider()).
				SetProviderSource(querier.ProviderSource()).
				SetCursor(*stats.latestCursor).
				SetLastSyncedAt(time.Now().UTC()).
				OnConflict(sql.ConflictColumns(pesc.FieldProvider, pesc.FieldProviderSource)).
				Update(func(u *ent.ProviderEventSyncCursorUpsert) {
					u.UpdateCursor()
					u.UpdateLastSyncedAt()
					u.UpdateUpdatedAt()
				})
			if cursorErr := saveCursor.Exec(ctx); cursorErr != nil {
				return fmt.Errorf("failed to save cursor: %w", cursorErr)
			}
		}
		return nil
	}

	req := rez.ProviderEventQueryRequest{CursorAfter: cursorAfter}
	for result, pullErr := range querier.PullEvents(ctx, req) {
		if pullErr != nil {
			return stats, fmt.Errorf("pull provider events: %w", pullErr)
		}
		stats.pulled++

		syncItem := syncBatchItem{
			args:        processProviderEventArgs{Event: result.Event},
			cursorAfter: result.CursorAfter,
		}
		if validErr := syncItem.args.validate(); validErr != nil {
			return stats, fmt.Errorf("invalid event args: %w", validErr)
		}
		batch = append(batch)
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

type processProviderEventArgs struct {
	Event rez.ProviderEvent
}

func (processProviderEventArgs) Kind() string {
	return "process-provider-event"
}

func (a processProviderEventArgs) validate() error {
	if a.Event.Provider == "" {
		return fmt.Errorf("event provider is required")
	}
	if a.Event.ProviderSource == "" {
		return fmt.Errorf("event provider_source is required")
	}
	if a.Event.SubjectRef == "" {
		return fmt.Errorf("event subject_ref is required")
	}
	if len(a.Event.Payload) == 0 {
		return fmt.Errorf("event payload is required")
	}
	if a.Event.ProviderEventRef == "" {
		return fmt.Errorf("event provider_delivery_ref is required")
	}
	return nil
}

type processProviderEventResult struct {
	processTime    time.Duration
	processSuccess bool
	normalizeCount int

	projectionSuccess bool
	projectionTime    time.Duration
}

func (s *ProviderEventService) processProviderEvent(ctx context.Context, ev rez.ProviderEvent) (*processProviderEventResult, error) {
	if _, tenantOk := execution.GetContext(ctx).TenantID(); !tenantOk {
		return nil, fmt.Errorf("tenant not found in context")
	}

	processStart := time.Now()
	proc, procExists := s.lookupEventProcessor(ev.Provider)
	if !procExists {
		return nil, fmt.Errorf("no provider event processor registered for provider %s", ev.Provider)
	}

	normalizedEvents, procErr := proc.Process(ctx, ev)

	slog.DebugContext(ctx, "processed provider event",
		"provider", ev.Provider,
		"source", ev.ProviderSource,
		"subject_ref", ev.SubjectRef,
		"error", procErr,
		"normalized_count", len(normalizedEvents),
	)

	res := &processProviderEventResult{}
	res.processTime = time.Since(processStart)
	res.normalizeCount = len(normalizedEvents)
	if procErr != nil {
		return res, fmt.Errorf("processing provider event: %w", procErr)
	}
	res.processSuccess = true

	saveNormalizedEventsFn := func(tx *ent.Tx) error {
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
				SetAttributes(ev.Attributes)

			c.OnConflict(sql.ConflictColumns(
				ne.FieldProvider,
				ne.FieldProviderSource,
				ne.FieldProviderEventRef,
				ne.FieldKind,
				ne.FieldSubjectRef,
			)).UpdateNewValues()
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

		numDuplicate := 0
		for _, r := range res {
			if r.UniqueSkippedAsDuplicate {
				numDuplicate++
			}
		}
		slog.DebugContext(ctx, "inserted projection job for normalized events",
			"count", len(res),
			"numDuplicate", numDuplicate,
		)

		return nil
	}

	if len(normalizedEvents) == 0 {
		if saveErr := ent.WithTx(ctx, s.db, saveNormalizedEventsFn); saveErr != nil {
			return res, fmt.Errorf("saving normalized events: %w", saveErr)
		}
	}
	return res, nil
}

func (s *ProviderEventService) projectNormalizedEvent(ctx context.Context, ev *ent.NormalizedEvent) error {
	failed := map[string]error{}
	for name, handlerFn := range eventprojections.GetHandlers() {
		// query for existing projection status
		queryStatus := s.db.NormalizedEventProjectionStatus.Query().
			Where(neps.NormalizedEventID(ev.ID), neps.HandlerName(name))
		status, statusErr := queryStatus.Only(ctx)
		if statusErr != nil && !ent.IsNotFound(statusErr) {
			failed[name] = fmt.Errorf("query projection status: %w", statusErr)
			continue
		}
		if status != nil {
			// don't project if this is in progress or already succeeded
			if status.Status == neps.StatusSucceeded || status.Status == neps.StatusPending {
				continue
			}
		} else {
			setPendingStatus := s.db.NormalizedEventProjectionStatus.Create().
				SetStatus(neps.StatusPending).
				SetNormalizedEventID(ev.ID).
				SetHandlerName(name).
				SetLastAttemptedAt(time.Now().UTC())
			status, statusErr = setPendingStatus.Save(ctx)
			if statusErr != nil {
				failed[name] = fmt.Errorf("mark projection pending: %w", statusErr)
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
			failed[name] = fmt.Errorf("update projection status: %w", statusErr)
			continue
		}
	}

	for name, err := range failed {
		slog.WarnContext(ctx, "failed to update projection status",
			"name", name,
			"err", err,
		)
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
