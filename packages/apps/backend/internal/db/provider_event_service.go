package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/jobs"
	"github.com/rezible/rezible/telemetry"
)

type ProviderEventService struct {
	logger *slog.Logger
	db     *ent.Client

	jobService rez.JobsService
	metrics    *providerEventMetrics

	processorsMu sync.RWMutex
	processors   map[string]rez.ProviderEventProcessor
}

func NewProviderEventService(ctx context.Context, db *ent.Client, js rez.JobsService) *ProviderEventService {
	pe := &ProviderEventService{
		logger:       telemetry.NewLogger(ctx, telemetry.WithLogPackage("provider_events")),
		db:           db,
		jobService:   js,
		metrics:      newProviderEventMetrics(),
		processorsMu: sync.RWMutex{},
		processors:   make(map[string]rez.ProviderEventProcessor),
	}
	jobs.RegisterWorkerFunc(pe.ProcessProviderEvent)

	return pe
}

func (s *ProviderEventService) RegisterEventProcessor(provider, providerSource string, proc rez.ProviderEventProcessor) {
	s.setEventProcessor(provider, providerSource, proc)
}

func (s *ProviderEventService) Ingest(ctx context.Context, ev rez.ProviderEvent) error {
	res, ingestErr := s.ingest(ctx, ev)
	s.metrics.recordIngested(ctx, ev.Provider, ev.ProviderSource, res, ingestErr)
	return ingestErr
}

type ingestProviderEventResult struct {
	duplicate bool
}

func (s *ProviderEventService) ingest(ctx context.Context, ev rez.ProviderEvent) (*ingestProviderEventResult, error) {
	processEvent := processProviderEventArgs{
		Provider:                 strings.TrimSpace(ev.Provider),
		ProviderSource:           strings.TrimSpace(ev.ProviderSource),
		SubjectRef:               strings.TrimSpace(ev.SubjectRef),
		ReceivedAt:               ev.ReceivedAt,
		ProviderEventDeliveryRef: ev.ProviderDeliveryRef,
		ContentType:              ev.ContentType,
		RequestMetadata:          ev.RequestMetadata,
		Payload:                  ev.Payload,
	}
	if processEvent.ReceivedAt.IsZero() {
		processEvent.ReceivedAt = time.Now().UTC()
	}

	if processEvent.Provider == "" {
		return nil, fmt.Errorf("provider event provider is required")
	}
	if processEvent.ProviderSource == "" {
		return nil, fmt.Errorf("provider event provider_source is required")
	}
	if processEvent.SubjectRef == "" {
		return nil, fmt.Errorf("provider event subject_ref is required")
	}
	if len(processEvent.Payload) == 0 {
		return nil, fmt.Errorf("provider event payload is required")
	}

	if processEvent.ProviderEventDeliveryRef == "" {
		hash := sha256.New()
		hash.Write([]byte(strings.TrimSpace(processEvent.Provider)))
		hash.Write([]byte{0})
		hash.Write([]byte(strings.TrimSpace(processEvent.ProviderSource)))
		hash.Write([]byte{0})
		hash.Write(processEvent.Payload)
		processEvent.ProviderEventDeliveryRef = hex.EncodeToString(hash.Sum(nil))
	}

	res, insertErr := s.jobService.Insert(ctx, processEvent, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{ByArgs: true},
	})
	if insertErr != nil {
		return nil, fmt.Errorf("could not insert provider event job: %w", insertErr)
	}

	var duplicate bool
	if res != nil && res.UniqueSkippedAsDuplicate {
		duplicate = true
		s.logger.Debug("skipped duplicate provider event",
			"provider", ev.Provider,
			"source", ev.ProviderSource,
		)
	}
	return &ingestProviderEventResult{duplicate: duplicate}, nil
}

func (s *ProviderEventService) setEventProcessor(provider, providerSource string, proc rez.ProviderEventProcessor) {
	s.processorsMu.Lock()
	defer s.processorsMu.Unlock()
	s.processors[s.makeEventProcessorKey(provider, providerSource)] = proc
}

func (s *ProviderEventService) makeEventProcessorKey(provider, providerSource string) string {
	return fmt.Sprintf("%s:%s", strings.ToLower(strings.TrimSpace(provider)), strings.TrimSpace(providerSource))
}

func (s *ProviderEventService) lookupEventProcessor(provider, providerSource string) (rez.ProviderEventProcessor, bool) {
	s.processorsMu.Lock()
	defer s.processorsMu.Unlock()
	proc, ok := s.processors[s.makeEventProcessorKey(provider, providerSource)]
	return proc, ok
}

type processProviderEventArgs struct {
	Provider                 string            `json:"provider"`
	ProviderSource           string            `json:"provider_source"`
	SubjectRef               string            `json:"subject_ref"`
	ReceivedAt               time.Time         `json:"received_at"`
	Payload                  []byte            `json:"payload"`
	ContentType              string            `json:"content_type,omitempty"`
	RequestMetadata          map[string]string `json:"request_metadata,omitempty"`
	ProviderEventDeliveryRef string            `json:"provider_event_delivery_ref,omitempty" river:"unique"`
}

func (processProviderEventArgs) Kind() string {
	return "process-provider-event"
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
	provEvent := rez.ProviderEvent{
		Provider:            args.Provider,
		ProviderSource:      args.ProviderSource,
		SubjectRef:          args.SubjectRef,
		ReceivedAt:          args.ReceivedAt,
		Payload:             args.Payload,
		ContentType:         args.ContentType,
		RequestMetadata:     args.RequestMetadata,
		ProviderDeliveryRef: args.ProviderEventDeliveryRef,
	}
	normalizedEvents, procErr := proc.Process(ctx, provEvent)
	res.processTime = time.Since(processStart)
	res.normalizeCount = len(normalizedEvents)
	if procErr != nil {
		return res, fmt.Errorf("processing provider event: %w", procErr)
	}
	res.processSuccess = true

	if res.normalizeCount > 0 {
		projectionStart := time.Now()
		tenantId := normalizedEvents[0].TenantID
		for i := 1; i < res.normalizeCount; i++ {
			if normalizedEvents[i].TenantID != tenantId {
				return res, fmt.Errorf("multiple tenant ids in events")
			}
		}
		projErr := s.projectNormalizedEvents(execution.SystemTenantContext(ctx, tenantId), normalizedEvents)
		res.projectionTime = time.Since(projectionStart)
		if projErr != nil {
			return res, fmt.Errorf("saving normalized events: %w", projErr)
		}
		res.projectionSuccess = true
	}

	return res, nil
}

func (s *ProviderEventService) projectNormalizedEvents(ctx context.Context, normalizedEvents ent.NormalizedEvents) error {
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
