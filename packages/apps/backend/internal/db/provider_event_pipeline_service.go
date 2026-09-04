package db

import (
	"context"
	stdsql "database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	nep "github.com/rezible/rezible/ent/normalizedeventprojection"
	nepe "github.com/rezible/rezible/ent/normalizedeventprojectionentity"
	"github.com/rezible/rezible/pkg/jobs"
)

type ProviderEventPipelineService struct {
	db         rez.Database
	jobs       rez.JobService
	logger     *slog.Logger
	telemetry  *providerEventTelemetry
	processors map[string]rez.ProviderEventProcessor
	projection rez.EventProjectionService
}

func NewProviderEventPipelineService(ts rez.TelemetryService, db rez.Database, jobSvc rez.JobService, processors map[string]rez.ProviderEventProcessor, projection rez.EventProjectionService) (*ProviderEventPipelineService, error) {
	logger := ts.NewLogger(rez.NewLoggerOptions{Name: "provider_events"})
	pe := &ProviderEventPipelineService{
		db:         db,
		jobs:       jobSvc,
		logger:     logger,
		telemetry:  newProviderEventTelemetry(ts, logger),
		processors: processors,
		projection: projection,
	}
	return pe, nil
}

func (s *ProviderEventPipelineService) Ingest(ctx context.Context, ev rez.ProviderEvent) error {
	duplicate, ingestErr := s.queueIngest(ctx, ev)
	s.telemetry.recordIngested(ctx, ev, duplicate, ingestErr)
	return ingestErr
}

func (s *ProviderEventPipelineService) queueIngest(ctx context.Context, ev rez.ProviderEvent) (bool, error) {
	if err := s.validateProviderEvent(ev); err != nil {
		return false, err
	}

	args := ProcessProviderEventArgs{Event: ev}
	insertOpts := args.InsertOpts()
	jobRes, insertErr := s.jobs.Insert(ctx, args, &insertOpts)
	if insertErr != nil {
		return false, fmt.Errorf("could not insert provider event job: %w", insertErr)
	}
	return jobRes.UniqueSkippedAsDuplicate, nil
}

func (s *ProviderEventPipelineService) SyncEvents(ctx context.Context, querier rez.ProviderEventQuerier, sourceCursors rez.ProviderEventSourceCursors) rez.ProviderEventSyncResult {
	res := rez.ProviderEventSyncResult{
		SourceCursorsAfter:  sourceCursors,
		SourceSyncDurations: make(map[string]time.Duration),
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
				Args:       ProcessProviderEventArgs{Event: item.Event},
				InsertOpts: &river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}},
			}
		}
		results, insertErr := s.jobs.InsertMany(ctx, params)
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
				if batchItem.ProviderEventSourceCursorAfter != nil {
					res.SourceCursorsAfter[batchItem.Event.ProviderEventSource] = *batchItem.ProviderEventSourceCursorAfter
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
		if eventErr := s.validateProviderEvent(result.Event); eventErr != nil {
			res.SyncErrors = append(res.SyncErrors, eventErr)
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
		if result.ProviderEventSourceCursorAfter == nil {
			break
		}
	}
	flushBatch()
	return res
}

func (s *ProviderEventPipelineService) validateProviderEvent(ev rez.ProviderEvent) error {
	eventRef := rez.ProviderResourceRef{
		Provider:          ev.Provider,
		ProviderNamespace: ev.ProviderNamespace,
		ResourceRef:       "event",
	}
	if refErr := eventRef.Validate(); refErr != nil {
		return fmt.Errorf("invalid event provider identity: %w", refErr)
	}
	if ev.ProviderEventSource == "" {
		return fmt.Errorf("event provider_event_source is required")
	}
	if len(ev.Attributes) == 0 {
		return fmt.Errorf("event attributes are required")
	}
	if ev.ProviderEventRef == "" {
		return fmt.Errorf("event provider_event_ref is required")
	}
	if ev.ReceivedAt.IsZero() {
		return fmt.Errorf("event received_at is required")
	}
	return nil
}

type processProviderEventResult struct {
	error error

	processTime    time.Duration
	processSuccess bool
	normalizeCount int
}

func (s *ProviderEventPipelineService) processProviderEvent(ctx context.Context, prov rez.ProviderEvent) processProviderEventResult {
	var res processProviderEventResult

	proc, ok := s.processors[prov.Provider]
	if !ok {
		res.error = fmt.Errorf("no event processor registered for provider '%s'", prov.Provider)
		return res
	}

	processStart := time.Now()
	procEvents, procErr := proc.ProcessProviderEvent(ctx, prov)
	res.processTime = time.Since(processStart)
	res.normalizeCount = len(procEvents)
	res.processSuccess = procErr == nil

	if procErr != nil {
		res.error = fmt.Errorf("processing event: %w", procErr)
		return res
	}

	if saveErr := s.saveNormalizedEvents(ctx, procEvents); saveErr != nil {
		res.error = fmt.Errorf("saving normalized events: %w", saveErr)
	}

	return res
}

func (s *ProviderEventPipelineService) saveNormalizedEvents(ctx context.Context, evts ent.NormalizedEvents) error {
	if len(evts) == 0 {
		return nil
	}

	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		ids := make([]uuid.UUID, 0, len(evts))
		for _, ev := range evts {
			id := ev.ID
			if id == uuid.Nil {
				id = uuid.New()
			}
			create := tx.NormalizedEvent.Create().
				SetID(id).
				SetKind(ev.Kind).
				SetProvider(ev.Provider).
				SetProviderNamespace(ev.ProviderNamespace).
				SetProviderResourceRef(ev.ProviderResourceRef).
				SetProviderEventSource(ev.ProviderEventSource).
				SetProviderEventRef(ev.ProviderEventRef).
				SetOccurredAt(ev.OccurredAt).
				SetReceivedAt(ev.ReceivedAt).
				SetAttributes(ev.Attributes)
			insert := create.OnConflict(sql.ConflictColumns(
				ne.FieldTenantID,
				ne.FieldProvider,
				ne.FieldProviderNamespace,
				ne.FieldProviderEventSource,
				ne.FieldProviderEventRef,
				ne.FieldProviderResourceRef,
			)).DoNothing()
			_, insertErr := insert.ID(ctx)
			if insertErr != nil {
				if errors.Is(insertErr, stdsql.ErrNoRows) {
					continue
				}
				return fmt.Errorf("insert normalized event: %w", insertErr)
			}
			ids = append(ids, id)
		}

		if len(ids) > 0 {
			params := make([]river.InsertManyParams, len(ids))
			for i, id := range ids {
				params[i] = river.InsertManyParams{
					Args: jobs.ProjectNormalizedEvent{EventId: id},
				}
			}
			_, jobErr := s.jobs.InsertMany(ctx, params)
			if jobErr != nil {
				return fmt.Errorf("inserting project events: %w", jobErr)
			}
			slog.Debug("inserted projection jobs", "new", len(ids))
		}
		return nil
	})
}

func (s *ProviderEventPipelineService) projectNormalizedEvent(ctx context.Context, ev *ent.NormalizedEvent) error {
	projectorFunc, ok := s.projection.GetEventProjectorFunc(ev)
	if !ok {
		return nil
	}

	projectEventFn := func(ctx context.Context, ev *ent.NormalizedEvent) (projectedEntities []rez.ProjectedEntityRef, err error) {
		defer func() {
			if v := recover(); v != nil {
				err = fmt.Errorf("projector panic: %v", v)
			}
		}()
		projectedEntities, err = projectorFunc(ctx, ev)
		return projectedEntities, err
	}

	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, "normalized_event_projection", ev.ID.String()); lockErr != nil {
			return fmt.Errorf("lock normalized event projection: %w", lockErr)
		}

		queryProjection := tx.NormalizedEventProjection.Query().
			Where(nep.EventID(ev.ID))
		projected, queryErr := queryProjection.Exist(ctx)
		if queryErr != nil {
			return fmt.Errorf("query projection receipt: %w", queryErr)
		} else if projected {
			return nil
		}

		projectedEntities, projectorErr := projectEventFn(ctx, ev)
		if projectorErr != nil {
			return projectorErr
		}

		createProjection := tx.NormalizedEventProjection.Create().
			SetEventID(ev.ID)
		proj, saveProjErr := createProjection.Save(ctx)
		if saveProjErr != nil {
			return fmt.Errorf("create projection receipt: %w", saveProjErr)
		}

		if len(projectedEntities) == 0 {
			return nil
		}

		createRefs := tx.NormalizedEventProjectionEntity.
			MapCreateBulk(projectedEntities, func(c *ent.NormalizedEventProjectionEntityCreate, i int) {
				ref := projectedEntities[i]
				c.SetProjectionID(proj.ID).
					SetDomainEntityKind(ref.Kind).
					SetDomainEntityID(ref.Id)
			})

		upsertRefs := createRefs.
			OnConflictColumns(nepe.FieldTenantID, nepe.FieldProjectionID, nepe.FieldDomainEntityID).
			DoNothing()
		if refsErr := upsertRefs.Exec(ctx); refsErr != nil {
			return fmt.Errorf("create projection entities: %w", refsErr)
		}
		return nil
	})
}

type providerEventTelemetry struct {
	logger           *slog.Logger
	ingested         metric.Int64Counter
	processed        metric.Int64Counter
	processSeconds   metric.Float64Histogram
	normalizedEvents metric.Int64Counter
}

func newProviderEventTelemetry(ts rez.TelemetryService, logger *slog.Logger) *providerEventTelemetry {
	meter := ts.DefaultMeter()
	processSeconds, processSecondsErr := meter.Float64Histogram("rezible.backend.provider_events.normalize_duration", metric.WithDescription("Provider event normalization processing duration"), metric.WithUnit("s"))
	ingested, ingestedErr := meter.Int64Counter("rezible.backend.provider_events.ingested", metric.WithDescription("Provider events ingested"))
	processed, processedErr := meter.Int64Counter("rezible.backend.provider_events.processed", metric.WithDescription("Provider events processed"))
	normalizedEvents, normalizedEventsErr := meter.Int64Counter("rezible.backend.provider_events.normalized_events", metric.WithDescription("Normalized provider events saved"))
	telErr := errors.Join(processSecondsErr, ingestedErr, processedErr, normalizedEventsErr)
	if telErr != nil {
		panic("telemetry instruments err: " + telErr.Error())
	}
	return &providerEventTelemetry{
		logger:           logger,
		ingested:         ingested,
		processed:        processed,
		processSeconds:   processSeconds,
		normalizedEvents: normalizedEvents,
	}
}

func (m *providerEventTelemetry) recordIngested(ctx context.Context, ev rez.ProviderEvent, duplicate bool, err error) {
	if m == nil {
		return
	}
	m.ingested.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", ev.Provider),
		attribute.String("provider_event_source", ev.ProviderEventSource),
		attribute.Bool("success", err == nil),
		attribute.Bool("duplicate", duplicate),
	))
	if duplicate {
		m.logger.Info("skipped ingesting duplicate provider event",
			"provider", ev.Provider,
			"provider_event_source", ev.ProviderEventSource,
		)
	}
}

func (m *providerEventTelemetry) recordProcessed(ctx context.Context, ev rez.ProviderEvent, res processProviderEventResult) {
	if m == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("provider", ev.Provider),
		attribute.String("provider_event_source", ev.ProviderEventSource),
		attribute.Bool("success", res.error == nil),
		attribute.Bool("process_success", res.processSuccess),
	}
	m.processed.Add(ctx, 1, metric.WithAttributes(attrs...))

	logAttrs := []slog.Attr{
		slog.Any("provider", ev.Provider),
		slog.Any("provider_event_source", ev.ProviderEventSource),
		slog.Any("error", res.error),
	}
	if res.error == nil {
		logAttrs = append(logAttrs, slog.Any("normalized_count", res.normalizeCount))
		m.processSeconds.Record(ctx, res.processTime.Seconds(), metric.WithAttributes(attrs...))
		if res.normalizeCount > 0 {
			m.normalizedEvents.Add(ctx, int64(res.normalizeCount), metric.WithAttributes(attrs...))
		}
	}

	m.logger.LogAttrs(ctx, slog.LevelInfo, "processed provider event", logAttrs...)
}
