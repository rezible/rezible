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
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/jobs"
	riverqueue "github.com/riverqueue/river"
)

type ProviderEventService struct {
	logger     *slog.Logger
	db         *ent.Client
	jobService rez.JobsService

	processorsMu sync.RWMutex
	processors   map[string]rez.ProviderEventProcessor
}

func NewProviderEventService(db *ent.Client, js rez.JobsService) *ProviderEventService {
	pe := &ProviderEventService{
		logger:     slog.Default(),
		db:         db,
		jobService: js,
		processors: make(map[string]rez.ProviderEventProcessor),
	}

	jobs.RegisterWorkerFunc(pe.processProviderEvent)

	return pe
}

func (s *ProviderEventService) Ingest(ctx context.Context, ev rez.ProviderEvent) error {
	processEvent := processProviderEvent{
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

	if validErr := s.validateProviderEvent(processEvent); validErr != nil {
		return fmt.Errorf("invalid event received: %w", validErr)
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

	res, insertErr := s.jobService.Insert(ctx, processEvent, &riverqueue.InsertOpts{
		UniqueOpts: riverqueue.UniqueOpts{ByArgs: true},
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

func (s *ProviderEventService) validateProviderEvent(e processProviderEvent) error {
	if strings.TrimSpace(e.Provider) == "" {
		return fmt.Errorf("provider event provider is required")
	}
	if strings.TrimSpace(e.Source) == "" {
		return fmt.Errorf("provider event source is required")
	}
	if len(e.Payload) == 0 {
		return fmt.Errorf("provider event payload is required")
	}
	return nil
}

func (s *ProviderEventService) RegisterEventProcessor(prov, src string, proc rez.ProviderEventProcessor) {
	if proc == nil {
		return
	}
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

type processProviderEvent struct {
	Provider        string            `json:"provider"`
	Source          string            `json:"source"`
	ReceivedAt      time.Time         `json:"received_at"`
	Payload         []byte            `json:"payload"`
	ContentType     string            `json:"content_type,omitempty"`
	RequestMetadata map[string]string `json:"request_metadata,omitempty"`
	DedupeKey       string            `json:"dedupe_key,omitempty" river:"unique"`
}

func (processProviderEvent) Kind() string {
	return "process-provider-event"
}

func (e processProviderEvent) getEvent() rez.ProviderEvent {
	return rez.ProviderEvent{
		Provider:        e.Provider,
		Source:          e.Source,
		ReceivedAt:      e.ReceivedAt,
		Payload:         e.Payload,
		ContentType:     e.ContentType,
		RequestMetadata: e.RequestMetadata,
		DedupeKey:       e.DedupeKey,
	}
}

func (s *ProviderEventService) processProviderEvent(ctx context.Context, prov processProviderEvent) error {
	proc, procExists := s.lookupEventProcessor(prov.Provider, prov.Source)
	if !procExists {
		return fmt.Errorf("no provider event processor registered for %s %s", prov.Provider, prov.Source)
	}
	normalizedEvents, procErr := proc.Process(ctx, prov.getEvent())
	if procErr != nil {
		return fmt.Errorf("processing provider event: %w", procErr)
	}
	if len(normalizedEvents) == 0 {
		return nil
	}
	tenantId := normalizedEvents[0].TenantID
	for i := 1; i < len(normalizedEvents)-1; i++ {
		if normalizedEvents[i].TenantID != tenantId {
			return fmt.Errorf("multiple tenant ids in events")
		}
	}
	ctx = execution.SystemTenantContext(ctx, tenantId)
	if saveErr := s.saveAndProjectNormalizedEvents(ctx, normalizedEvents); saveErr != nil {
		return fmt.Errorf("saving normalized events: %w", saveErr)
	}
	return nil
}

func (s *ProviderEventService) saveAndProjectNormalizedEvents(ctx context.Context, normalizedEvents ent.NormalizedEvents) error {
	upsertCols := sql.ConflictColumns(
		normalizedevent.FieldTenantID,
		normalizedevent.FieldProvider,
		normalizedevent.FieldSource,
		normalizedevent.FieldProcessingVersion,
		normalizedevent.FieldSourceEventKey,
		normalizedevent.FieldKind,
		normalizedevent.FieldSubjectExternalRef,
	)

	mapEvent := func(c *ent.NormalizedEventCreate, i int) {
		ev := normalizedEvents[i]
		c.SetTenantID(ev.TenantID).
			SetProvider(ev.Provider).
			SetSource(ev.Source).
			SetKind(ev.Kind).
			SetSubjectKind(ev.SubjectKind).
			SetSubjectExternalRef(ev.SubjectExternalRef).
			SetSourceEventKey(ev.SourceEventKey).
			SetOccurredAt(ev.OccurredAt).
			SetReceivedAt(ev.ReceivedAt).
			SetProcessingVersion(ev.ProcessingVersion).
			SetAttributes(ev.Attributes).
			SetDedupeKey(ev.DedupeKey).
			OnConflict(upsertCols).UpdateNewValues()
	}

	persistEventsTx := func(tx *ent.Tx) error {
		evs, upsertErr := tx.NormalizedEvent.MapCreateBulk(normalizedEvents, mapEvent).Save(ctx)
		if upsertErr != nil {
			return fmt.Errorf("upsert normalized events: %w", upsertErr)
		}

		projectEvents := make([]riverqueue.InsertManyParams, len(evs))
		for i, ev := range evs {
			projectEvents[i] = riverqueue.InsertManyParams{
				Args: projectNormalizedEvent{EventId: ev.ID},
				InsertOpts: &riverqueue.InsertOpts{
					UniqueOpts: riverqueue.UniqueOpts{ByArgs: true},
				},
			}
		}
		if _, jobErr := s.jobService.InsertManyTx(ctx, tx, projectEvents); jobErr != nil {
			return fmt.Errorf("inserting project events: %w", jobErr)
		}

		return nil
	}
	if txErr := ent.WithTx(ctx, s.db, persistEventsTx); txErr != nil {
		return fmt.Errorf("persist normalized events: %w", txErr)
	}
	return nil
}

type projectNormalizedEvent struct {
	EventId uuid.UUID
}

func (projectNormalizedEvent) Kind() string {
	return "project-normalized-event"
}

func (s *ProviderEventService) projectNormalizedEvent(ctx context.Context, args projectNormalizedEvent) error {
	ev, queryErr := s.db.NormalizedEvent.Get(ctx, args.EventId)
	if queryErr != nil {
		return fmt.Errorf("query events: %w", queryErr)
	}
	switch ev.Kind {
	case "chat":
		return s.projectChatMessagePosted(ctx, ev)
	}
	return nil
}

func (s *ProviderEventService) projectChatMessagePosted(ctx context.Context, ev *ent.NormalizedEvent) error {
	channelID, _ := ev.Attributes["channel_id"].(string)
	text, _ := ev.Attributes["text"].(string)
	if channelID == "" {
		return fmt.Errorf("chat message normalized event missing channel_id attribute")
	}

	inc, incErr := s.db.Incident.Query().Where(incident.ChatChannelID(channelID)).First(ctx)
	if incErr != nil {
		if ent.IsNotFound(incErr) {
			return nil
		}
		return fmt.Errorf("lookup incident by slack channel: %w", incErr)
	}

	// TODO: incident channel message
	slog.DebugContext(ctx, "projected incident channel chat message",
		"incidentId", inc.ID.String(),
		"text", text,
	)
	return nil
}
