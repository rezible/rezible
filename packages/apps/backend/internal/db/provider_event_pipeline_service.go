package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	neps "github.com/rezible/rezible/ent/normalizedeventprojectionstatus"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/jobs"
	"github.com/rezible/rezible/projections"
)

type ProviderEventPipelineService struct {
	logger     *slog.Logger
	db         rez.Database
	reg        *projections.PipelineRegistry
	jobService rez.JobService
	telemetry  *providerEventTelemetry
}

func NewProviderEventPipelineService(ts rez.TelemetryService, db rez.Database, jobSvc rez.JobService, reg *projections.PipelineRegistry) (*ProviderEventPipelineService, error) {
	logger := ts.NewLogger(rez.NewLoggerOptions{PackageName: "provider_events"})
	pe := &ProviderEventPipelineService{
		logger:     logger,
		db:         db,
		jobService: jobSvc,
		reg:        reg,
		telemetry:  newProviderEventTelemetry(ts, logger),
	}
	jobs.RegisterWorkerFunc(pe.HandleProcessEventJob)
	jobs.RegisterWorkerFunc(pe.HandleEventProjectionJob)
	return pe, nil
}

func (s *ProviderEventPipelineService) Ingest(ctx context.Context, ev rez.ProviderEvent) error {
	duplicate, ingestErr := s.ingest(ctx, ev)
	s.telemetry.recordIngested(ctx, ev, duplicate, ingestErr)
	return ingestErr
}

func (s *ProviderEventPipelineService) ingest(ctx context.Context, ev rez.ProviderEvent) (bool, error) {
	if ev.Provider == "" {
		return false, fmt.Errorf("event provider is required")
	} else if ev.ProviderSource == "" {
		return false, fmt.Errorf("event provider_source is required")
	} else if ev.ProviderSubjectRef == "" {
		return false, fmt.Errorf("event subject_ref is required")
	} else if len(ev.Payload) == 0 {
		return false, fmt.Errorf("event payload is required")
	} else if ev.ProviderEventRef == "" {
		return false, fmt.Errorf("event provider_delivery_ref is required")
	}

	args := processProviderEventArgs{Event: ev}
	insertOpts := &river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}}
	insertRes, insertErr := s.jobService.Insert(ctx, args, insertOpts)
	if insertErr != nil {
		return false, fmt.Errorf("could not insert provider event job: %w", insertErr)
	}
	return insertRes.UniqueSkippedAsDuplicate, nil
}

type processProviderEventArgs struct {
	Event rez.ProviderEvent
}

func (processProviderEventArgs) Kind() string {
	return "process-provider-event"
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

	res, resErr := s.projectNormalizedEvent(ctx, ev)
	if resErr != nil {
		return fmt.Errorf("project event: %w", resErr)
	}

	retryableErrors, fatalErrors := res.errors()
	for projName, err := range fatalErrors {
		s.logger.ErrorContext(ctx, "fatal event projection error",
			"error", err.Error(),
			"name", projName,
			"ref", ev.ProviderEventRef,
		)
	}
	return errors.Join(retryableErrors...)
}

func (s *ProviderEventPipelineService) SyncEvents(ctx context.Context, querier rez.ProviderEventQuerier, sourceCursors map[string]string) rez.ProviderEventSyncResult {
	res := rez.ProviderEventSyncResult{
		SourceCursorsAfter: sourceCursors,
	}

	const batchSize = 100

	batch := make([]rez.ProviderEventQueryResult, 0, batchSize)

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
			res.SyncErrors = append(res.SyncErrors, fmt.Errorf("inserting process provider event jobs: %w", insertErr))
			return false
		}
		for i, result := range results {
			if result != nil && result.UniqueSkippedAsDuplicate {
				res.NumDuplicates++
			} else {
				res.EventsIngested++
			}
			if i < len(batch) {
				batchItem := batch[i]
				if batchItem.SourceCursorAfter != nil {
					res.SourceCursorsAfter[batchItem.Event.ProviderSource] = *batchItem.SourceCursorAfter
				}
			}
		}
		return true
	}

	for result, pullErr := range querier.QueryProviderEvents(ctx, sourceCursors) {
		if pullErr != nil {
			res.SyncErrors = append(res.SyncErrors, fmt.Errorf("pulling events from querier: %w", pullErr))
			break
		}
		if result == nil {
			break
		}
		res.EventsPulled++
		batch = append(batch, *result)
		if len(batch) >= batchSize {
			if flushOk := flushBatch(); !flushOk {
				break
			}
			batch = make([]rez.ProviderEventQueryResult, 0, batchSize)
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

func (s *ProviderEventPipelineService) processProviderEvent(ctx context.Context, prov rez.ProviderEvent) processProviderEventResult {
	var res processProviderEventResult
	if _, tenantOk := execution.GetContext(ctx).TenantID(); !tenantOk {
		res.error = fmt.Errorf("tenant not found in context")
		return res
	}

	processStart := time.Now()
	proc, ok := s.reg.GetProviderEventProcessor(prov.Provider)
	if !ok {
		res.error = fmt.Errorf("no event processors registered for provider '%s'", prov.Provider)
		return res
	}

	normalizedEvents, procErr := proc.ProcessProviderEvent(ctx, prov)

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

func (s *ProviderEventPipelineService) projectNormalizedEvent(ctx context.Context, ev *ent.NormalizedEvent) (*eventProjectionResult, error) {
	res := eventProjectionResult{handlerErrors: make(map[string][]error)}

	projStatuses, projStatusErr := s.getEventProjectionStatuses(ctx, ev.ID)
	if projStatusErr != nil {
		return nil, fmt.Errorf("query status: %w", projStatusErr)
	}

	for name, projector := range s.reg.GetEventProjectorsForKind(projections.SubjectKind(ev.SubjectKind)) {
		currStatus, statusExists := projStatuses[name]
		if statusExists && currStatus.state == neps.StatusSucceeded {
			continue
		}

		status, setPendingErr := s.setEventProjectionStatusPending(ctx, currStatus.id, name, ev.ID)
		if setPendingErr != nil {
			res.addHandlerError(name, fmt.Errorf("set status pending: %w", setPendingErr))
		}

		projErr := s.runEventProjector(ctx, ev, projector)
		if projErr != nil {
			res.addHandlerError(name, projErr)
		}

		if resErr := s.setEventProjectionStatusResult(ctx, status.ID, projErr); resErr != nil {
			res.addHandlerError(name, fmt.Errorf("set projection result status: %w", setPendingErr))
		}
	}

	return &res, nil
}

func (s *ProviderEventPipelineService) runEventProjector(ctx context.Context, ev *ent.NormalizedEvent, projector rez.NormalizedEventProjector) error {
	return s.db.WithTx(ctx, func(ctx context.Context, _ *ent.Client) (err error) {
		defer func() {
			if v := recover(); v != nil {
				slog.WarnContext(ctx, "event projection panic",
					"error", fmt.Sprintf("%+v", v),
					"stack", string(debug.Stack()),
					"event", ev.ID.String())
				err = fmt.Errorf("projector panic: %v", v)
			}
		}()
		return projector.HandleEventProjection(ctx, ev)
	})
}

type eventProjectionResult struct {
	handlerErrors map[string][]error
}

func (r eventProjectionResult) errors() ([]error, map[string]error) {
	retryable := make([]error, 0, len(r.handlerErrors))
	fatal := make(map[string]error)
	for name, handlerErrs := range r.handlerErrors {
		for _, handlerErr := range handlerErrs {
			if projections.IsRetryable(handlerErr) {
				retryable = append(retryable, fmt.Errorf("%s: %w", name, handlerErr))
			} else {
				fatal[name] = errors.Join(fatal[name], handlerErr)
			}
		}
	}
	return retryable, fatal
}

func (r eventProjectionResult) addHandlerError(name string, err error) {
	r.handlerErrors[name] = append(r.handlerErrors[name], err)
}

type eventProjectionStatus struct {
	id    uuid.UUID
	state neps.Status
}

func (s *ProviderEventPipelineService) getEventProjectionStatuses(ctx context.Context, eventId uuid.UUID) (map[string]eventProjectionStatus, error) {
	queryStatuses := s.db.Client(ctx).NormalizedEventProjectionStatus.Query().Where(neps.NormalizedEventID(eventId))
	statuses, queryStatusErr := queryStatuses.All(ctx)
	if queryStatusErr != nil {
		return nil, fmt.Errorf("query status: %w", queryStatusErr)
	}
	projStatuses := map[string]eventProjectionStatus{}
	for _, status := range statuses {
		projStatuses[status.HandlerName] = eventProjectionStatus{id: status.ID, state: status.Status}
	}
	return projStatuses, nil
}

func (s *ProviderEventPipelineService) setEventProjectionStatusPending(ctx context.Context, statusId uuid.UUID, projName string, evId uuid.UUID) (*ent.NormalizedEventProjectionStatus, error) {
	return s.setEventProjectionStatus(ctx, statusId, func(m *ent.NormalizedEventProjectionStatusMutation) {
		m.ClearFailedAt()
		m.ClearLastError()
		m.SetNormalizedEventID(evId)
		m.SetHandlerName(projName)
		m.SetStatus(neps.StatusPending)
		m.SetLastAttemptedAt(time.Now().UTC())
	})
}

func (s *ProviderEventPipelineService) setEventProjectionStatusResult(ctx context.Context, statusId uuid.UUID, projErr error) error {
	_, setErr := s.setEventProjectionStatus(ctx, statusId, func(m *ent.NormalizedEventProjectionStatusMutation) {
		if projErr == nil {
			m.ClearFailedAt()
			m.ClearLastError()
			m.SetStatus(neps.StatusSucceeded)
			m.SetSucceededAt(time.Now().UTC())
		} else {
			m.SetStatus(neps.StatusFailed)
			m.SetFailedAt(time.Now().UTC())
			m.SetLastError(projErr.Error())
		}
	})
	return setErr
}

func (s *ProviderEventPipelineService) setEventProjectionStatus(ctx context.Context, statusId uuid.UUID, setFn func(*ent.NormalizedEventProjectionStatusMutation)) (*ent.NormalizedEventProjectionStatus, error) {
	var mutator ent.EntityMutator[*ent.NormalizedEventProjectionStatus, *ent.NormalizedEventProjectionStatusMutation]
	if statusId != uuid.Nil {
		mutator = s.db.Client(ctx).NormalizedEventProjectionStatus.UpdateOneID(statusId)
	} else {
		mutator = s.db.Client(ctx).NormalizedEventProjectionStatus.Create()
	}
	setFn(mutator.Mutation())
	return mutator.Save(ctx)
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

func (m *providerEventTelemetry) recordIngested(ctx context.Context, ev rez.ProviderEvent, duplicate bool, err error) {
	if m == nil {
		return
	}
	m.ingested.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", ev.Provider),
		attribute.String("provider_source", ev.ProviderSource),
		attribute.Bool("success", err == nil),
		attribute.Bool("duplicate", duplicate),
	))
	if duplicate {
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
