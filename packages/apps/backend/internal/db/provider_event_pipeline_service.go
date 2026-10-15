package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/riverqueue/river"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/jobs"
)

type ProviderEventPipelineService struct {
	db        rez.Database
	jobs      rez.JobService
	ingester  rez.ProviderEventIngester
	telemetry *providerEventTelemetry
}

func NewProviderEventPipelineService(db rez.Database, jobSvc rez.JobService, ingester rez.ProviderEventIngester) (*ProviderEventPipelineService, error) {
	pe := &ProviderEventPipelineService{
		db:        db,
		jobs:      jobSvc,
		ingester:  ingester,
		telemetry: newProviderEventTelemetry(),
	}
	return pe, nil
}

func (s *ProviderEventPipelineService) Ingest(ctx context.Context, ev rez.ProviderEvent) error {
	duplicate, ingestErr := s.queueIngest(ctx, ev)
	s.telemetry.recordIngested(ctx, ev, duplicate, ingestErr)
	return ingestErr
}

func (s *ProviderEventPipelineService) queueIngest(ctx context.Context, ev rez.ProviderEvent) (bool, error) {
	if validateErr := s.validateProviderEvent(ev); validateErr != nil {
		return false, validateErr
	}
	args := &jobs.ProcessProviderEventArgs{Event: ev}
	insertOpts := args.InsertOpts()
	jobRes, insertErr := s.jobs.Insert(ctx, args, &insertOpts)
	if insertErr != nil {
		return false, fmt.Errorf("could not insert provider event job: %w", insertErr)
	}
	return jobRes.UniqueSkippedAsDuplicate, nil
}

// IngestMany validates every event before queueing any, then queues them in one insert so a delivery is
// queued together or not at all.
func (s *ProviderEventPipelineService) IngestMany(ctx context.Context, evs []rez.ProviderEvent) error {
	if len(evs) == 0 {
		return nil
	}
	params := make([]river.InsertManyParams, len(evs))
	for i, ev := range evs {
		if validateErr := s.validateProviderEvent(ev); validateErr != nil {
			return fmt.Errorf("event %d: %w", i, validateErr)
		}
		args := &jobs.ProcessProviderEventArgs{Event: ev}
		insertOpts := args.InsertOpts()
		params[i] = river.InsertManyParams{Args: args, InsertOpts: &insertOpts}
	}
	results, insertErr := s.jobs.InsertMany(ctx, params)
	for i, ev := range evs {
		duplicate := insertErr == nil && i < len(results) && results[i] != nil && results[i].UniqueSkippedAsDuplicate
		s.telemetry.recordIngested(ctx, ev, duplicate, insertErr)
	}
	if insertErr != nil {
		return fmt.Errorf("could not insert provider event jobs: %w", insertErr)
	}
	return nil
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
		slog.DebugContext(ctx, "flushing batch", "len", len(batch))
		params := make([]river.InsertManyParams, len(batch))
		for i, item := range batch {
			params[i] = river.InsertManyParams{
				Args:       &jobs.ProcessProviderEventArgs{Event: item.Event},
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

type providerEventTelemetry struct {
	ingested         metric.Int64Counter
	processed        metric.Int64Counter
	processSeconds   metric.Float64Histogram
	normalizedEvents metric.Int64Counter
}

func newProviderEventTelemetry() *providerEventTelemetry {
	meter := otel.Meter("github.com/rezible/rezible")
	processSeconds, processSecondsErr := meter.Float64Histogram("rezible.backend.provider_events.normalize_duration", metric.WithDescription("Provider event normalization processing duration"), metric.WithUnit("s"))
	ingested, ingestedErr := meter.Int64Counter("rezible.backend.provider_events.ingested", metric.WithDescription("Provider events ingested"))
	processed, processedErr := meter.Int64Counter("rezible.backend.provider_events.processed", metric.WithDescription("Provider events processed"))
	normalizedEvents, normalizedEventsErr := meter.Int64Counter("rezible.backend.provider_events.normalized_events", metric.WithDescription("Normalized provider events saved"))
	telErr := errors.Join(processSecondsErr, ingestedErr, processedErr, normalizedEventsErr)
	if telErr != nil {
		panic("telemetry instruments err: " + telErr.Error())
	}
	return &providerEventTelemetry{
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
		slog.InfoContext(ctx, "skipped ingesting duplicate provider event",
			"provider", ev.Provider,
			"provider_event_source", ev.ProviderEventSource,
		)
	}
}

func (m *providerEventTelemetry) recordProcessed(ctx context.Context, ev rez.ProviderEvent, duration time.Duration, savedCount int, err error) {
	if m == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("provider", ev.Provider),
		attribute.String("provider_event_source", ev.ProviderEventSource),
		attribute.Bool("success", err == nil),
	}
	m.processed.Add(ctx, 1, metric.WithAttributes(attrs...))

	logAttrs := []slog.Attr{
		slog.Any("provider", ev.Provider),
		slog.Any("provider_event_source", ev.ProviderEventSource),
		slog.Any("error", err),
	}
	if err == nil {
		logAttrs = append(logAttrs, slog.Any("normalized_count", savedCount))
		m.processSeconds.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
		if savedCount > 0 {
			m.normalizedEvents.Add(ctx, int64(savedCount), metric.WithAttributes(attrs...))
		}
	}

	slog.LogAttrs(ctx, slog.LevelInfo, "processed provider event", logAttrs...)
}
