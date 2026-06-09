package projections

import (
	"reflect"
	"sync"

	rez "github.com/rezible/rezible"
)

type PipelineRegistry struct {
	eventProcessors   map[string]rez.ProviderEventProcessor
	eventProcessorsMu sync.RWMutex

	eventProjectors   map[SubjectKind]map[string]rez.NormalizedEventProjector
	eventProjectorsMu sync.RWMutex
}

func NewPipelineRegistry() *PipelineRegistry {
	return &PipelineRegistry{
		eventProcessors: make(map[string]rez.ProviderEventProcessor),
		eventProjectors: make(map[SubjectKind]map[string]rez.NormalizedEventProjector),
	}
}

func (r *PipelineRegistry) RegisterProviderEventProcessors(processor rez.ProviderEventProcessor, providers ...string) {
	r.eventProcessorsMu.Lock()
	defer r.eventProcessorsMu.Unlock()
	for _, provider := range providers {
		//slog.Debug("registered provider event processor", "provider", provider)
		r.eventProcessors[provider] = processor
	}
}

func (r *PipelineRegistry) GetProviderEventProcessor(provider string) (rez.ProviderEventProcessor, bool) {
	proc, ok := r.eventProcessors[provider]
	return proc, ok
}

func (r *PipelineRegistry) RegisterEventProjector(handler rez.NormalizedEventProjector, kinds ...SubjectKind) {
	r.eventProjectorsMu.Lock()
	defer r.eventProjectorsMu.Unlock()

	name := reflect.TypeOf(handler).String()
	for _, kind := range kinds {
		if _, exists := r.eventProjectors[kind]; !exists {
			r.eventProjectors[kind] = make(map[string]rez.NormalizedEventProjector)
		}
		r.eventProjectors[kind][name] = handler
	}
	//slog.Debug("registered event projection handler", "name", name, "kinds", kinds)
}

func (r *PipelineRegistry) GetEventProjectorsForKind(kind SubjectKind) map[string]rez.NormalizedEventProjector {
	return r.eventProjectors[kind]
}
