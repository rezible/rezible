package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/pkg/projections"
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
	logger := ts.NewLogger(rez.NewLoggerOptions{PackageName: "provider_events"})
	pe := &ProviderEventPipelineService{
		db:         db,
		jobs:       jobSvc,
		logger:     logger,
		telemetry:  newProviderEventTelemetry(ts, logger),
		processors: processors,
		projection: projection,
	}
	jobs.RegisterWorkerFunc(pe.HandleProcessEventJob)
	jobs.RegisterWorkerFunc(pe.HandleEventProjectionJob)
	return pe, nil
}

func (s *ProviderEventPipelineService) Ingest(ctx context.Context, ev rez.ProviderEvent) error {
	duplicate, ingestErr := s.queueIngest(ctx, ev)
	s.telemetry.recordIngested(ctx, ev, duplicate, ingestErr)
	return ingestErr
}

func (s *ProviderEventPipelineService) queueIngest(ctx context.Context, ev rez.ProviderEvent) (bool, error) {
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
	insertOpts := args.InsertOpts()
	jobRes, insertErr := s.jobs.Insert(ctx, args, &insertOpts)
	if insertErr != nil {
		return false, fmt.Errorf("could not insert provider event job: %w", insertErr)
	}
	return jobRes.UniqueSkippedAsDuplicate, nil
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

	if projectionErr != nil {
		if s.db.IsTransientError(projectionErr) {
			return projections.Retryable(projectionErr)
		}
		if projectionErr == nil || projections.IsRetryable(projectionErr) {
			return projectionErr
		}
		s.logger.ErrorContext(ctx, "fatal event projection error",
			"error", projectionErr.Error(),
			"ref", ev.ProviderEventRef,
		)
		return river.JobCancel(projectionErr)
	}

	return nil
}

func (s *ProviderEventPipelineService) SyncEvents(ctx context.Context, querier rez.ProviderEventQuerier, sourceCursors rez.ProviderEventQuerySourceCursors) rez.ProviderEventSyncResult {
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

	processTime    time.Duration
	processSuccess bool
	normalizeCount int
}

func (s *ProviderEventPipelineService) processProviderEvent(ctx context.Context, prov rez.ProviderEvent) processProviderEventResult {
	var res processProviderEventResult

	proc, ok := s.processors[prov.Provider]
	if !ok {
		res.error = fmt.Errorf("no event processors registered for provider '%s'", prov.Provider)
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

var normalizedEventUniqueColumns = sql.ConflictColumns(
	ne.FieldTenantID,
	ne.FieldProvider,
	ne.FieldProviderSource,
	ne.FieldProviderEventRef,
	ne.FieldProviderSubjectRef,
)

func (s *ProviderEventPipelineService) saveNormalizedEvents(ctx context.Context, evts ent.NormalizedEvents) error {
	if len(evts) == 0 {
		return nil
	}

	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		ids := make([]uuid.UUID, len(evts))
		for i, ev := range evts {
			ids[i] = ev.ID
			if ids[i] == uuid.Nil {
				ids[i] = uuid.New()
			}
		}

		insertBulk := tx.NormalizedEvent.MapCreateBulk(evts, func(c *ent.NormalizedEventCreate, i int) {
			ev := evts[i]
			c.SetID(ids[i]).
				SetProvider(ev.Provider).
				SetProviderSource(ev.ProviderSource).
				SetProviderEventRef(ev.ProviderEventRef).
				SetKind(ev.Kind).
				SetSubjectKind(ev.SubjectKind).
				SetProviderSubjectRef(ev.ProviderSubjectRef).
				SetOccurredAt(ev.OccurredAt).
				SetReceivedAt(ev.ReceivedAt).
				SetAttributes(ev.Attributes)
		})

		insertBulk.OnConflict(normalizedEventUniqueColumns).
			DoNothing()

		if insertErr := insertBulk.Exec(ctx); insertErr != nil {
			return fmt.Errorf("insert normalized events: %w", insertErr)
		}

		if len(ids) > 0 {
			params := make([]river.InsertManyParams, len(ids))
			for i, id := range ids {
				params[i] = river.InsertManyParams{
					Args: jobs.ProjectNormalizedEvent{EventId: id},
				}
			}
			res, jobErr := s.jobs.InsertMany(ctx, params)
			if jobErr != nil {
				return fmt.Errorf("inserting project events: %w", jobErr)
			}
			dups := 0
			for _, r := range res {
				if r.UniqueSkippedAsDuplicate {
					dups++
				}
			}
			slog.Debug("inserted projection jobs", "duplicates", dups, "new", len(ids)-dups)
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
		logAttrs = append(logAttrs, slog.Any("normalized_count", res.normalizeCount))
		m.processSeconds.Record(ctx, res.processTime.Seconds(), metric.WithAttributes(attrs...))
		if res.normalizeCount > 0 {
			m.normalizedEvents.Add(ctx, int64(res.normalizeCount), metric.WithAttributes(attrs...))
		}
	}

	m.logger.LogAttrs(ctx, slog.LevelInfo, "processed provider event", logAttrs...)
}
