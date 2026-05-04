package river

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/jobs"
	riverqueue "github.com/riverqueue/river"
)

type ProviderEventIngestor struct {
	logger     *slog.Logger
	jobService *JobService
}

func NewProviderEventIngestor(s *JobService) *ProviderEventIngestor {
	pe := &ProviderEventIngestor{
		logger:     slog.Default(),
		jobService: s,
	}

	jobs.RegisterWorkerFunc(pe.processProviderEvent)

	return pe
}

func (s *ProviderEventIngestor) IngestEvent(ctx context.Context, ev rez.ProviderEvent) error {
	processEvent := processProviderEvent{
		Provider:        strings.TrimSpace(ev.Provider),
		Source:          strings.TrimSpace(ev.Source),
		ReceivedAt:      ev.ReceivedAt,
		Payload:         ev.Payload,
		ContentType:     ev.ContentType,
		RequestMetadata: ev.RequestMetadata,
		DedupeKey:       ev.DedupeKey,
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

func (s *ProviderEventIngestor) validateProviderEvent(e processProviderEvent) error {
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

func (s *ProviderEventIngestor) RegisterEventProcessor(processor rez.ProviderEventProcessor) {
	providerEventProcessors.register(processor)
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

func (s *ProviderEventIngestor) processProviderEvent(ctx context.Context, ev processProviderEvent) error {
	processor := providerEventProcessors.get(ev.Provider, ev.Source)
	if processor == nil {
		return fmt.Errorf("no provider event processor registered for %s %s", ev.Provider, ev.Source)
	}
	return processor.ProcessProviderEvent(ctx, ev.getEvent())
}

type providerEventProcessorRegistry struct {
	mu         sync.RWMutex
	processors map[string]rez.ProviderEventProcessor
}

var providerEventProcessors = &providerEventProcessorRegistry{
	processors: make(map[string]rez.ProviderEventProcessor),
}

func (r *providerEventProcessorRegistry) makeKey(provider, source string) string {
	return fmt.Sprintf("%s:%s", strings.ToLower(strings.TrimSpace(provider)), strings.TrimSpace(source))
}

func (r *providerEventProcessorRegistry) register(processor rez.ProviderEventProcessor) {
	if processor == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.processors[r.makeKey(processor.Provider(), processor.Source())] = processor
}

func (r *providerEventProcessorRegistry) get(provider, source string) rez.ProviderEventProcessor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.processors[r.makeKey(provider, source)]
}

func (r *providerEventProcessorRegistry) clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.processors = make(map[string]rez.ProviderEventProcessor)
}
