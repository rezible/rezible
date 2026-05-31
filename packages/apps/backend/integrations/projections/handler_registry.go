package projections

import (
	"context"
	"sync"

	"github.com/rezible/rezible/ent"
)

type EventProjectionHandlerFunc = func(context.Context, *ent.Client, *ent.NormalizedEvent) error

type EventProjectionHandler struct {
	Handler      EventProjectionHandlerFunc
	SubjectKinds []SubjectKind
}

type EventProjectionHandlerRegistry struct {
	projectionFuncs   map[string]EventProjectionHandler
	projectionFuncsMu sync.RWMutex
}

func NewEventProjectionHandlerRegistry() *EventProjectionHandlerRegistry {
	return &EventProjectionHandlerRegistry{
		projectionFuncs: make(map[string]EventProjectionHandler),
	}
}

func (r *EventProjectionHandlerRegistry) RegisterHandler(name string, handler EventProjectionHandlerFunc, subjectKinds ...SubjectKind) {
	r.projectionFuncsMu.Lock()
	defer r.projectionFuncsMu.Unlock()
	r.projectionFuncs[name] = EventProjectionHandler{
		Handler:      handler,
		SubjectKinds: subjectKinds,
	}
}

func (r *EventProjectionHandlerRegistry) GetHandlersFor(ev *ent.NormalizedEvent) map[string]EventProjectionHandlerFunc {
	r.projectionFuncsMu.RLock()
	defer r.projectionFuncsMu.RUnlock()

	handlers := make(map[string]EventProjectionHandlerFunc)
	if ev == nil {
		return handlers
	}
	for name, registered := range r.projectionFuncs {
		if len(registered.SubjectKinds) == 0 {
			handlers[name] = registered.Handler
			continue
		}
		for _, subjectKind := range registered.SubjectKinds {
			if subjectKind.Matches(ev) {
				handlers[name] = registered.Handler
				break
			}
		}
	}
	return handlers
}
