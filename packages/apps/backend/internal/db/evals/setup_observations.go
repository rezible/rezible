package evals

import (
	"context"

	rez "github.com/rezible/rezible"
)

// SetupObservations wraps the observation service for agent evaluations. An observation notified while a
// scenario's setup event is ingested is recorded against that event instead of being placed by detection,
// since detection does not place an observation a situation has and the scenario's setup decides which. Other
// notifications pass through.
type SetupObservations struct {
	next rez.SituationObservationService
}

func NewSetupObservations(next rez.SituationObservationService) *SetupObservations {
	return &SetupObservations{next: next}
}

func (s *SetupObservations) NotifyObservationChanged(ctx context.Context, ref rez.SituationObservationRef) error {
	if recorder, isSetup := ctx.Value(setupEventKey{}).(*setupEvent); isSetup {
		recorder.record(ref)
		return nil
	}
	return s.next.NotifyObservationChanged(ctx, ref)
}

func (s *SetupObservations) NotifyStructureChanged(ctx context.Context) error {
	return s.next.NotifyStructureChanged(ctx)
}

type setupEventKey struct{}

// setupEvent collects the observations notified while one setup event is ingested. Projection notifies on the
// ingesting goroutine.
type setupEvent struct {
	observations []rez.SituationObservationRef
}

func (e *setupEvent) record(ref rez.SituationObservationRef) {
	e.observations = append(e.observations, ref)
}

// withSetupEvent marks the context as ingesting a setup event; observations notified under it are recorded.
func withSetupEvent(ctx context.Context) (context.Context, *setupEvent) {
	recorder := &setupEvent{}
	return context.WithValue(ctx, setupEventKey{}, recorder), recorder
}
