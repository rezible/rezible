package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"sync"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	iesr "github.com/rezible/rezible/ent/integrationeventsyncrun"
	"github.com/riverqueue/river"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	neps "github.com/rezible/rezible/ent/normalizedeventprojectionstatus"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/integrations/projections"
	"github.com/rezible/rezible/jobs"
)

type ProviderEventService struct {
	logger *slog.Logger
	db     rez.Database

	eventProjectors   map[string]eventProjector
	eventProjectorsMu sync.RWMutex

	jobService   rez.JobService
	integrations rez.IntegrationService
	telemetry    *providerEventTelemetry
}

type eventProjector struct {
	handler      rez.EventProjectionHandler
	subjectKinds mapset.Set[projections.SubjectKind]
}

func NewProviderEventService(ts rez.TelemetryService, db rez.Database, jobSvc rez.JobService, intgs rez.IntegrationService) (*ProviderEventService, error) {
	logger := ts.NewLogger(rez.NewLoggerOptions{PackageName: "provider_events"})
	pe := &ProviderEventService{
		logger:          logger,
		db:              db,
		eventProjectors: make(map[string]eventProjector),
		jobService:      jobSvc,
		integrations:    intgs,
		telemetry:       newProviderEventTelemetry(ts, logger),
	}
	jobs.RegisterWorkerFunc(pe.HandleSyncIntegrationEventsJob)
	jobs.RegisterWorkerFunc(pe.HandleProcessEventJob)
	jobs.RegisterWorkerFunc(pe.HandleEventProjectionJob)
	return pe, nil
}

func (s *ProviderEventService) RegisterProjectionHandler(handler rez.EventProjectionHandler, kinds ...projections.SubjectKind) {
	s.eventProjectorsMu.Lock()
	defer s.eventProjectorsMu.Unlock()
	name, _ := strings.CutPrefix(reflect.TypeOf(handler).String(), "*")
	slog.Debug("registered event projection handler", "name", name, "kinds", kinds)
	s.eventProjectors[name] = eventProjector{
		handler:      handler,
		subjectKinds: mapset.NewSet(kinds...),
	}
}

func (s *ProviderEventService) Ingest(ctx context.Context, ev rez.ProviderEvent) error {
	res := s.ingest(ctx, ev)
	s.telemetry.recordIngested(ctx, ev, res)
	return res.error
}

type providerEventIngestResult struct {
	duplicate bool
	error     error
}

func (s *ProviderEventService) ingest(ctx context.Context, ev rez.ProviderEvent) providerEventIngestResult {
	var res providerEventIngestResult
	processArgs := processProviderEventArgs{Event: ev}
	if validationErr := processArgs.validate(); validationErr != nil {
		res.error = fmt.Errorf("invalid event: %w", validationErr)
		return res
	}
	insertOpts := &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{ByArgs: true},
	}
	insertRes, insertErr := s.jobService.Insert(ctx, processArgs, insertOpts)
	if insertErr != nil {
		res.error = fmt.Errorf("could not insert provider event job: %w", insertErr)
	} else {
		res.duplicate = insertRes.UniqueSkippedAsDuplicate
	}
	return res
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
	if a.Event.ProviderSubjectRef == "" {
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

func (s *ProviderEventService) HandleProcessEventJob(ctx context.Context, args processProviderEventArgs) error {
	res := s.processProviderEvent(ctx, args.Event)
	s.telemetry.recordProcessed(ctx, args.Event, res)
	return res.error
}

func (s *ProviderEventService) HandleEventProjectionJob(ctx context.Context, args jobs.ProjectNormalizedEvent) error {
	ev, queryErr := s.db.Client(ctx).NormalizedEvent.Get(ctx, args.EventId)
	if queryErr != nil {
		return fmt.Errorf("query normalized event: %w", queryErr)
	}

	res := s.projectNormalizedEvent(ctx, ev)

	for name, errs := range res.handlerErrors {
		s.logger.WarnContext(ctx, "failed to update projection status",
			"handler", name,
			"errors", errors.Join(errs...),
		)
	}

	return nil
}

func (s *ProviderEventService) HandleSyncIntegrationEventsJob(ctx context.Context, args jobs.SyncIntegrationEventsArgs) error {
	if args.IntegrationId == uuid.Nil {
		// TODO: sync all installed?
		return nil
	}

	intg, intgErr := s.integrations.GetInstalled(ctx, args.IntegrationId)
	if intgErr != nil {
		return fmt.Errorf("get installed integration: %w", intgErr)
	}

	querier, querierErr := s.integrations.GetProviderEventQuerier(ctx, intg.Integration())
	if querierErr != nil {
		return fmt.Errorf("get provider event querier: %w", querierErr)
	}

	sourceCursors := map[string]string{}
	for _, src := range args.Sources {
		// TODO: look up cursors from last sync
		sourceCursors[src] = ""
	}

	return s.SyncIntegrationEvents(ctx, rez.IntegrationEventSyncOptions{
		SyncReason:    args.SyncReason,
		Querier:       querier,
		SourceCursors: sourceCursors,
	})
}

func (s *ProviderEventService) SyncIntegrationEvents(ctx context.Context, opts rez.IntegrationEventSyncOptions) error {
	res := s.syncIntegrationEvents(ctx, opts.Querier, opts.SourceCursors)
	s.saveSyncIntegrationEventsResult(ctx, opts.SyncReason, res)
	return res.syncError
}

type integrationEventSyncResult struct {
	syncError error

	integrationId  uuid.UUID
	sourceCursors  map[string]string
	startedAt      time.Time
	eventsPulled   int
	eventsIngested int
	duplicates     int
}

func (s *ProviderEventService) syncIntegrationEvents(ctx context.Context, querier rez.IntegrationEventQuerier, sourceCursors map[string]string) integrationEventSyncResult {
	res := integrationEventSyncResult{
		startedAt:     time.Now(),
		integrationId: querier.Integration().ID,
		sourceCursors: map[string]string{},
	}

	const batchSize = 100

	batch := make([]rez.IntegrationEventQueryResult, 0, batchSize)

	flushBatch := func() bool {
		if len(batch) == 0 {
			return true
		}
		s.logger.DebugContext(ctx, "flushing batch", "len", len(batch))
		params := make([]river.InsertManyParams, len(batch))
		for i, item := range batch {
			params[i] = river.InsertManyParams{
				Args:       processProviderEventArgs{Event: item.Event},
				InsertOpts: &river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}},
			}
		}
		results, insertErr := s.jobService.InsertMany(ctx, params)
		if insertErr != nil {
			res.syncError = fmt.Errorf("inserting process provider event jobs: %w", insertErr)
			return false
		}
		for i, result := range results {
			if result != nil && result.UniqueSkippedAsDuplicate {
				res.duplicates++
			} else {
				res.eventsIngested++
			}
			if i < len(batch) {
				batchItem := batch[i]
				if batchItem.SourceCursorAfter != nil {
					res.sourceCursors[batchItem.Event.ProviderSource] = *batchItem.SourceCursorAfter
				}
			}
		}
		return true
	}

	for result, pullErr := range querier.PullEvents(ctx, sourceCursors) {
		if pullErr != nil {
			res.syncError = fmt.Errorf("pulling events from querier: %w", pullErr)
			break
		}
		if result == nil {
			break
		}
		res.eventsPulled++
		batch = append(batch, *result)
		if len(batch) >= batchSize {
			if flushOk := flushBatch(); !flushOk {
				break
			}
			batch = make([]rez.IntegrationEventQueryResult, 0, batchSize)
		}
		if result.SourceCursorAfter == nil {
			break
		}
	}
	flushBatch()
	return res
}

type processProviderEventResult struct {
	error error

	processTime                time.Duration
	processSuccess             bool
	normalizeCount             int
	insertProjectionDuplicates int
}

func (s *ProviderEventService) processProviderEvent(ctx context.Context, prov rez.ProviderEvent) processProviderEventResult {
	var res processProviderEventResult
	if _, tenantOk := execution.GetContext(ctx).TenantID(); !tenantOk {
		res.error = fmt.Errorf("tenant not found in context")
		return res
	}

	processStart := time.Now()
	proc, getProcErr := s.integrations.GetProviderEventProcessor(prov.Provider)
	if getProcErr != nil {
		res.error = fmt.Errorf("get '%s' processor: %w", prov.Provider, getProcErr)
		return res
	}

	normalizedEvents, procErr := proc.Process(ctx, prov)

	res.processTime = time.Since(processStart)
	res.normalizeCount = len(normalizedEvents)
	if procErr != nil {
		res.error = fmt.Errorf("processing event: %w", procErr)
		return res
	}
	res.processSuccess = true

	mapCreateEventFn := func(c *ent.NormalizedEventCreate, i int) {
		ev := normalizedEvents[i]
		c.SetProvider(ev.Provider).
			SetProviderSource(ev.ProviderSource).
			SetProviderEventRef(ev.ProviderEventRef).
			SetActivityKind(ev.ActivityKind).
			SetSubjectKind(ev.SubjectKind).
			SetProviderSubjectRef(ev.ProviderSubjectRef).
			SetOccurredAt(ev.OccurredAt).
			SetReceivedAt(ev.ReceivedAt).
			SetAttributes(ev.Attributes)
	}
	upsertConflictColumns := sql.ConflictColumns(
		ne.FieldTenantID,
		ne.FieldProvider,
		ne.FieldProviderSource,
		ne.FieldProviderEventRef,
		ne.FieldProviderSubjectRef,
	)
	eventRefs := mapset.NewSet[string]()
	for _, ev := range normalizedEvents {
		eventRefs.Add(ev.ProviderEventRef)
	}

	saveNormalizedEventsFn := func(txCtx context.Context, tx *ent.Client) error {
		upsertBulk := tx.NormalizedEvent.MapCreateBulk(normalizedEvents, mapCreateEventFn).
			OnConflict(upsertConflictColumns).
			UpdateNewValues()
		if upsertErr := upsertBulk.Exec(txCtx); upsertErr != nil {
			return fmt.Errorf("upsert normalized events: %w", upsertErr)
		}

		queryEvents := tx.NormalizedEvent.Query().
			Where(ne.ProviderEventRefIn(eventRefs.ToSlice()...))
		evs, evsErr := queryEvents.All(txCtx)
		if evsErr != nil {
			return fmt.Errorf("query normalized events: %w", evsErr)
		}

		params := make([]river.InsertManyParams, len(evs))
		for i, ev := range evs {
			params[i] = river.InsertManyParams{
				Args: jobs.ProjectNormalizedEvent{
					EventId: ev.ID,
				},
				InsertOpts: &river.InsertOpts{
					//UniqueOpts: river.UniqueOpts{ByArgs: true},
				},
			}
		}
		insertRes, jobErr := s.jobService.InsertMany(txCtx, params)
		if jobErr != nil {
			return fmt.Errorf("inserting project events: %w", jobErr)
		}
		for _, r := range insertRes {
			if r.UniqueSkippedAsDuplicate {
				res.insertProjectionDuplicates++
			}
		}
		return nil
	}

	if len(normalizedEvents) > 0 {
		if saveErr := s.db.WithTx(ctx, saveNormalizedEventsFn); saveErr != nil {
			res.error = fmt.Errorf("saving normalized events: %w", saveErr)
		}
	}
	return res
}

type projectNormalizedEventResult struct {
	handlerErrors map[string][]error
}

func (s *ProviderEventService) projectNormalizedEvent(ctx context.Context, ev *ent.NormalizedEvent) projectNormalizedEventResult {
	res := projectNormalizedEventResult{
		handlerErrors: make(map[string][]error),
	}
	appendHandlerErr := func(name string, err error) {
		res.handlerErrors[name] = append(res.handlerErrors[name], err)
	}
	kind := projections.SubjectKind(ev.SubjectKind)
	for name, proj := range s.eventProjectors {
		if !proj.subjectKinds.Contains(kind) {
			continue
		}
		// query for existing projection status
		queryStatus := s.db.Client(ctx).NormalizedEventProjectionStatus.Query().
			Where(neps.NormalizedEventID(ev.ID), neps.HandlerName(name))
		status, statusErr := queryStatus.Only(ctx)
		if statusErr != nil && !ent.IsNotFound(statusErr) {
			appendHandlerErr(name, fmt.Errorf("query projection status: %w", statusErr))
			continue
		}
		if status != nil {
			// don't project if this is in progress or already succeeded
			if status.Status == neps.StatusSucceeded || status.Status == neps.StatusPending {
				continue
			}
		} else {
			setPendingStatus := s.db.Client(ctx).NormalizedEventProjectionStatus.Create().
				SetStatus(neps.StatusPending).
				SetNormalizedEventID(ev.ID).
				SetHandlerName(name).
				SetLastAttemptedAt(time.Now().UTC())
			status, statusErr = setPendingStatus.Save(ctx)
			if statusErr != nil {
				appendHandlerErr(name, fmt.Errorf("mark projection pending: %w", statusErr))
				continue
			}
		}

		projectionErr := proj.handler.HandleEventProjection(ctx, ev)

		update := status.Update()
		if projectionErr == nil {
			update.SetStatus(neps.StatusSucceeded).
				SetSucceededAt(time.Now().UTC()).
				ClearFailedAt().
				ClearLastError()
		} else {
			appendHandlerErr(name, fmt.Errorf("save projection tx: %w", projectionErr))
			update.SetStatus(neps.StatusFailed).
				SetFailedAt(time.Now().UTC()).
				SetLastError(projectionErr.Error())
		}
		if statusErr = update.Exec(ctx); statusErr != nil {
			appendHandlerErr(name, fmt.Errorf("update projection status: %w", statusErr))
		}
	}

	return res
}

func (s *ProviderEventService) saveSyncIntegrationEventsResult(ctx context.Context, reason string, res integrationEventSyncResult) {
	saveResult := s.db.Client(ctx).IntegrationEventSyncRun.Create().
		SetSyncReason(reason).
		SetIntegrationID(res.integrationId).
		SetSourceCursors(res.sourceCursors).
		SetEventsPulled(res.eventsPulled).
		SetEventsIngested(res.eventsIngested).
		SetDuplicates(res.duplicates).
		SetStartedAt(res.startedAt).
		SetFinishedAt(time.Now().UTC()).
		SetStatus(iesr.StatusSuccess)
	if res.syncError != nil {
		saveResult.SetStatus(iesr.StatusFailed)
		saveResult.SetFailureMessage(res.syncError.Error())
	}
	if saveRunErr := saveResult.Exec(ctx); saveRunErr != nil {
		s.logger.ErrorContext(ctx, "failed to save provider events", "err", saveRunErr)
	}
}

type providerEventTelemetry struct {
	logger            *slog.Logger
	ingested          metric.Int64Counter
	processed         metric.Int64Counter
	processSeconds    metric.Float64Histogram
	projectionSeconds metric.Float64Histogram
	normalizedEvents  metric.Int64Counter
}

func newProviderEventTelemetry(ts rez.TelemetryService, logger *slog.Logger) *providerEventTelemetry {
	meter := ts.DefaultMeter()
	processSeconds, processSecondsErr := meter.Float64Histogram("rezible.backend.provider_events.normalize_duration", metric.WithDescription("Provider event normalization processing duration"), metric.WithUnit("s"))
	ingested, ingestedErr := meter.Int64Counter("rezible.backend.provider_events.ingested", metric.WithDescription("Provider events ingested"))
	processed, processedErr := meter.Int64Counter("rezible.backend.provider_events.processed", metric.WithDescription("Provider events processed"))
	normalizedEvents, normalizedEventsErr := meter.Int64Counter("rezible.backend.provider_events.normalized_events", metric.WithDescription("Normalized provider events saved"))
	projectionSeconds, projectionSecondsErr := meter.Float64Histogram("rezible.backend.provider_events.projection_duration", metric.WithDescription("Normalized event projection duration"), metric.WithUnit("s"))
	telErr := errors.Join(processSecondsErr, ingestedErr, processedErr, normalizedEventsErr, projectionSecondsErr)
	if telErr != nil {
		panic("telemetry instruments err: " + telErr.Error())
	}
	return &providerEventTelemetry{
		logger:            logger,
		ingested:          ingested,
		processed:         processed,
		processSeconds:    processSeconds,
		normalizedEvents:  normalizedEvents,
		projectionSeconds: projectionSeconds,
	}
}

func (m *providerEventTelemetry) recordIngested(ctx context.Context, ev rez.ProviderEvent, res providerEventIngestResult) {
	if m == nil {
		return
	}
	m.ingested.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", ev.Provider),
		attribute.String("provider_source", ev.ProviderSource),
		attribute.Bool("success", res.error == nil),
		attribute.Bool("duplicate", res.duplicate),
	))
	if res.duplicate {
		m.logger.Info("skipped ingesting duplicate provider event",
			"provider", ev.Provider,
			"source", ev.ProviderSource,
		)
	}
}

func (m *providerEventTelemetry) recordProcessed(ctx context.Context, ev rez.ProviderEvent, res processProviderEventResult) {
	if m == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("provider", ev.Provider),
		attribute.String("provider_source", ev.ProviderSource),
		attribute.Bool("success", res.error == nil),
		attribute.Bool("process_success", res.processSuccess),
	}
	m.processed.Add(ctx, 1, metric.WithAttributes(attrs...))

	logAttrs := []slog.Attr{
		slog.Any("provider", ev.Provider),
		slog.Any("source", ev.ProviderSource),
		slog.Any("subject_ref", ev.ProviderSubjectRef),
		slog.Any("error", res.error),
	}
	if res.error == nil {
		logAttrs = append(logAttrs,
			slog.Any("normalized_count", res.normalizeCount),
			slog.Any("insert_projection_duplicates", res.insertProjectionDuplicates))
		m.processSeconds.Record(ctx, res.processTime.Seconds(), metric.WithAttributes(attrs...))
		if res.normalizeCount > 0 {
			m.normalizedEvents.Add(ctx, int64(res.normalizeCount), metric.WithAttributes(attrs...))
		}
	}

	m.logger.LogAttrs(ctx, slog.LevelInfo, "processed provider event", logAttrs...)
}
