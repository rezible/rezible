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
)

type ProviderEventService struct {
	logger *slog.Logger
	db     *ent.Client

	jobService rez.JobsService

	processorsMu sync.RWMutex
	processors   map[string]rez.ProviderEventProcessor
}

func NewProviderEventService(db *ent.Client, js rez.JobsService) *ProviderEventService {
	pe := &ProviderEventService{
		logger:       slog.Default(),
		db:           db,
		jobService:   js,
		processorsMu: sync.RWMutex{},
		processors:   make(map[string]rez.ProviderEventProcessor),
	}
	jobs.RegisterWorkerFunc(pe.processProviderEvent)

	return pe
}

func (s *ProviderEventService) RegisterEventProcessor(prov, src string, proc rez.ProviderEventProcessor) {
	s.setEventProcessor(prov, src, proc)
}

func (s *ProviderEventService) Ingest(ctx context.Context, ev rez.ProviderEvent) error {
	processEvent := processProviderEventArgs{
		Provider:        strings.TrimSpace(ev.Provider),
		Source:          strings.TrimSpace(ev.Source),
		ReceivedAt:      time.Now().UTC(),
		DedupeKey:       ev.DedupeKey,
		RequestMetadata: ev.RequestMetadata,
		Payload:         ev.Payload,
	}
	if processEvent.ReceivedAt.IsZero() {
		processEvent.ReceivedAt = time.Now().UTC()
	}

	if processEvent.Provider == "" {
		return fmt.Errorf("provider event provider is required")
	}
	if processEvent.Source == "" {
		return fmt.Errorf("provider event source is required")
	}
	if len(processEvent.Payload) == 0 {
		return fmt.Errorf("provider event payload is required")
	}

	if processEvent.DedupeKey == "" {
		hash := sha256.New()
		hash.Write([]byte(strings.TrimSpace(processEvent.Provider)))
		hash.Write([]byte{0})
		hash.Write([]byte(strings.TrimSpace(processEvent.Source)))
		hash.Write([]byte{0})
		hash.Write(processEvent.Payload)
		processEvent.DedupeKey = hex.EncodeToString(hash.Sum(nil))
	}

	res, insertErr := s.jobService.Insert(ctx, processEvent, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{ByArgs: true},
	})
	if insertErr != nil {
		return fmt.Errorf("could not insert provider event job: %w", insertErr)
	}
	if res != nil && res.UniqueSkippedAsDuplicate {
		s.logger.Debug("skipped duplicate provider event",
			"provider", ev.Provider,
			"source", ev.Source,
		)
	}
	return nil
}

func (s *ProviderEventService) setEventProcessor(prov, src string, proc rez.ProviderEventProcessor) {
	s.processorsMu.Lock()
	defer s.processorsMu.Unlock()
	s.processors[s.makeEventProcessorKey(prov, src)] = proc
}

func (s *ProviderEventService) makeEventProcessorKey(provider, source string) string {
	return fmt.Sprintf("%s:%s", strings.ToLower(strings.TrimSpace(provider)), strings.TrimSpace(source))
}

func (s *ProviderEventService) lookupEventProcessor(provider, source string) (rez.ProviderEventProcessor, bool) {
	s.processorsMu.Lock()
	defer s.processorsMu.Unlock()
	proc, ok := s.processors[s.makeEventProcessorKey(provider, source)]
	return proc, ok
}

type processProviderEventArgs struct {
	Provider        string            `json:"provider"`
	Source          string            `json:"source"`
	ReceivedAt      time.Time         `json:"received_at"`
	Payload         []byte            `json:"payload"`
	ContentType     string            `json:"content_type,omitempty"`
	RequestMetadata map[string]string `json:"request_metadata,omitempty"`
	DedupeKey       string            `json:"dedupe_key,omitempty" river:"unique"`
}

func (processProviderEventArgs) Kind() string {
	return "process-provider-event"
}

func (s *ProviderEventService) processProviderEvent(ctx context.Context, args processProviderEventArgs) error {
	proc, procExists := s.lookupEventProcessor(args.Provider, args.Source)
	if !procExists {
		return fmt.Errorf("no provider event processor registered for %s %s", args.Provider, args.Source)
	}
	provEvent := rez.ProviderEvent{
		Provider:        args.Provider,
		Source:          args.Source,
		ReceivedAt:      args.ReceivedAt,
		Payload:         args.Payload,
		ContentType:     args.ContentType,
		RequestMetadata: args.RequestMetadata,
		DedupeKey:       args.DedupeKey,
	}
	normalizedEvents, procErr := proc.Process(ctx, provEvent)
	if procErr != nil {
		return fmt.Errorf("processing provider event: %w", procErr)
	}
	if len(normalizedEvents) == 0 {
		return nil
	}
	tenantId := normalizedEvents[0].TenantID
	for i := 1; i < len(normalizedEvents); i++ {
		if normalizedEvents[i].TenantID != tenantId {
			return fmt.Errorf("multiple tenant ids in events")
		}
	}
	tenantCtx := execution.SystemTenantContext(ctx, tenantId)
	if saveErr := s.saveAndProjectNormalizedEvents(tenantCtx, normalizedEvents); saveErr != nil {
		return fmt.Errorf("saving normalized events: %w", saveErr)
	}
	return nil
}

func (s *ProviderEventService) saveAndProjectNormalizedEvents(ctx context.Context, normalizedEvents ent.NormalizedEvents) error {
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
			SetOccurredAt(ev.OccurredAt).
			SetReceivedAt(ev.ReceivedAt).
			SetProcessingVersion(ev.ProcessingVersion).
			SetAttributes(ev.Attributes).
			SetDedupeKey(ev.DedupeKey).
			OnConflict(upsertCols).UpdateNewValues()
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
