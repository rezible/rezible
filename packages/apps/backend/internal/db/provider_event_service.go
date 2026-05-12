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
	"github.com/rezible/rezible/internal/projections"
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

func NewProviderEventService(ctx context.Context, db *ent.Client, js rez.JobsService, intgs rez.IntegrationsService) *ProviderEventService {
	pe := &ProviderEventService{
		logger:       telemetry.NewLogger(ctx, telemetry.WithLogPackage("provider_events")),
		db:           db,
		jobService:   js,
		integrations: intgs,
		metrics:      newProviderEventMetrics(),
		processorsMu: sync.RWMutex{},
		processors:   make(map[string]rez.ProviderEventProcessor),
	}
	jobs.RegisterWorkerFunc(pe.HandleProviderEventSyncJob)
	jobs.RegisterWorkerFunc(pe.ProcessProviderEvent)
	jobs.RegisterWorkerFunc(pe.ProjectNormalizedEvent)
	return pe
}

func (s *ProviderEventService) makeEventProcessorKey(provider, providerSource string) string {
	return fmt.Sprintf("%s:%s", strings.ToLower(strings.TrimSpace(provider)), strings.TrimSpace(providerSource))
}

func (s *ProviderEventService) RegisterEventProcessors(provider string, procs map[string]rez.ProviderEventProcessor) {
	s.processorsMu.Lock()
	defer s.processorsMu.Unlock()
	for source, processor := range procs {
		key := s.makeEventProcessorKey(provider, source)
		s.processors[key] = processor
	}
}

func (s *ProviderEventService) lookupEventProcessor(provider, providerSource string) (rez.ProviderEventProcessor, bool) {
	s.processorsMu.RLock()
	defer s.processorsMu.RUnlock()
	proc, ok := s.processors[s.makeEventProcessorKey(provider, providerSource)]
	return proc, ok
}

func (s *ProviderEventService) Ingest(ctx context.Context, ev rez.ProviderEvent) (*rez.ProviderEventIngestResult, error) {
	res, ingestErr := s.ingest(ctx, ev)
	s.metrics.recordIngested(ctx, ev.Provider, ev.ProviderSource, res, ingestErr)
	return res, ingestErr
}

func (s *ProviderEventService) ingest(ctx context.Context, ev rez.ProviderEvent) (*rez.ProviderEventIngestResult, error) {
	processEvent, validationErr := s.makeProcessEventArgs(ev)
	if validationErr != nil {
		return nil, fmt.Errorf("invalid event: %w", validationErr)
	}

	insertRes, insertErr := s.jobService.Insert(ctx, *processEvent, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{ByArgs: true},
	})
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
	Provider            string            `json:"provider"`
	ProviderSource      string            `json:"provider_source"`
	SubjectRef          string            `json:"subject_ref"`
	ReceivedAt          time.Time         `json:"received_at"`
	Payload             []byte            `json:"payload"`
	ContentType         string            `json:"content_type,omitempty"`
	RequestMetadata     map[string]string `json:"request_metadata,omitempty"`
	ProviderDeliveryRef string            `json:"provider_delivery_ref,omitempty" river:"unique"`
}

func (processProviderEventArgs) Kind() string {
	return "process-provider-event"
}

func (s *ProviderEventService) makeProcessEventArgs(ev rez.ProviderEvent) (*processProviderEventArgs, error) {
	args := processProviderEventArgs{
		Provider:            strings.TrimSpace(ev.Provider),
		ProviderSource:      strings.TrimSpace(ev.ProviderSource),
		ProviderDeliveryRef: strings.TrimSpace(ev.ProviderDeliveryRef),
		SubjectRef:          strings.TrimSpace(ev.SubjectRef),
		ContentType:         strings.TrimSpace(ev.ContentType),
		ReceivedAt:          ev.ReceivedAt,
		RequestMetadata:     ev.RequestMetadata,
		Payload:             ev.Payload,
	}
	if ev.ReceivedAt.IsZero() {
		args.ReceivedAt = time.Now().UTC()
	}
	if args.Provider == "" {
		return nil, fmt.Errorf("provider event provider is required")
	}
	if args.ProviderSource == "" {
		return nil, fmt.Errorf("provider event provider_source is required")
	}
	if args.SubjectRef == "" {
		return nil, fmt.Errorf("provider event subject_ref is required")
	}
	if len(args.Payload) == 0 {
		return nil, fmt.Errorf("provider event payload is required")
	}
	if args.ProviderDeliveryRef == "" {
		return nil, fmt.Errorf("provider event provider_delivery_ref is required")
	}
	return &args, nil
}

func (args processProviderEventArgs) toProviderEvent() rez.ProviderEvent {
	return rez.ProviderEvent{
		Provider:            args.Provider,
		ProviderSource:      args.ProviderSource,
		SubjectRef:          args.SubjectRef,
		ReceivedAt:          args.ReceivedAt,
		Payload:             args.Payload,
		ContentType:         args.ContentType,
		RequestMetadata:     args.RequestMetadata,
		ProviderDeliveryRef: args.ProviderDeliveryRef,
	}
}

func (s *ProviderEventService) ProcessProviderEvent(ctx context.Context, args processProviderEventArgs) error {
	res, err := s.processProviderEvent(ctx, args)
	s.metrics.recordProcessed(ctx, args.Provider, args.ProviderSource, res, err)
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
	res := &processProviderEventResult{}

	processStart := time.Now()
	proc, procExists := s.lookupEventProcessor(args.Provider, args.ProviderSource)
	if !procExists {
		return nil, fmt.Errorf("no provider event processor registered for %s %s", args.Provider, args.ProviderSource)
	}
	normalizedEvents, procErr := proc.Process(ctx, args.toProviderEvent())
	res.processTime = time.Since(processStart)
	res.normalizeCount = len(normalizedEvents)
	if procErr != nil {
		return res, fmt.Errorf("processing provider event: %w", procErr)
	}
	res.processSuccess = true

	if res.normalizeCount > 0 {
		projectionStart := time.Now()
		tenantID := normalizedEvents[0].TenantID
		for i := 1; i < res.normalizeCount; i++ {
			if normalizedEvents[i].TenantID != tenantID {
				return res, fmt.Errorf("multiple tenant ids in events")
			}
		}
		projErr := s.saveNormalizedEvents(execution.SystemTenantContext(ctx, tenantID), normalizedEvents)
		res.projectionTime = time.Since(projectionStart)
		if projErr != nil {
			return res, fmt.Errorf("saving normalized events: %w", projErr)
		}
		res.projectionSuccess = true
	}

	return res, nil
}

func (s *ProviderEventService) saveNormalizedEvents(ctx context.Context, normalizedEvents ent.NormalizedEvents) error {
	upsertCols := sql.ConflictColumns(
		ne.FieldTenantID,
		ne.FieldProvider,
		ne.FieldProviderSource,
		ne.FieldProcessingVersion,
		ne.FieldProviderEventRef,
		ne.FieldKind,
		ne.FieldSubjectRef,
	)

	mapEvent := func(c *ent.NormalizedEventCreate, i int) {
		ev := normalizedEvents[i]
		c.SetTenantID(ev.TenantID).
			SetProvider(ev.Provider).
			SetProviderSource(ev.ProviderSource).
			SetKind(ev.Kind).
			SetSubjectKind(ev.SubjectKind).
			SetSubjectRef(ev.SubjectRef).
			SetProviderEventRef(ev.ProviderEventRef).
			SetProviderEventDeliveryRef(ev.ProviderEventDeliveryRef).
			SetOccurredAt(ev.OccurredAt).
			SetReceivedAt(ev.ReceivedAt).
			SetProcessingVersion(ev.ProcessingVersion).
			SetAttributes(ev.Attributes)

		c.OnConflict(upsertCols).UpdateNewValues()
	}

	return ent.WithTx(ctx, s.db, func(tx *ent.Tx) error {
		evs, upsertErr := tx.NormalizedEvent.MapCreateBulk(normalizedEvents, mapEvent).Save(ctx)
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
		slog.DebugContext(ctx, "requested projection for normalized events", "count", len(res))

		return nil
	})
}

func (s *ProviderEventService) HandleProviderEventSyncJob(ctx context.Context, args jobs.ProviderEventSyncJob) error {
	if len(args.ProviderSources) == 0 {
		slog.WarnContext(ctx, "TODO: support syncing all provider events")
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
		return fmt.Errorf("finish provider event sync run: %w", updateErr)
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

	conflictCols := sql.ConflictColumns(
		pesc.FieldTenantID,
		pesc.FieldProvider,
		pesc.FieldProviderSource,
	)
	create.OnConflict(conflictCols).Update(func(u *ent.ProviderEventSyncCursorUpsert) {
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
	for name, handlerFn := range projections.GetEventProjectionHandlers(ev.Kind) {
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
