package evals

import (
	"context"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
)

// SetupSignals wraps the signal service for agent evaluations. A signal notified while a scenario's setup event
// is ingested is recorded against that event instead of being placed by detection, since a signal belongs to
// at most one situation and the scenario's setup decides which. Other notifications pass through.
type SetupSignals struct {
	next rez.SituationSignalService
}

func NewSetupSignals(next rez.SituationSignalService) *SetupSignals {
	return &SetupSignals{next: next}
}

func (s *SetupSignals) NotifySignalChanged(ctx context.Context, signalEntityID uuid.UUID) error {
	if recorder, isSetup := ctx.Value(setupEventKey{}).(*setupEvent); isSetup {
		recorder.record(signalEntityID)
		return nil
	}
	return s.next.NotifySignalChanged(ctx, signalEntityID)
}

type setupEventKey struct{}

// setupEvent collects the signals notified while one setup event is ingested. Projection notifies on the ingesting goroutine.
type setupEvent struct {
	signals []uuid.UUID
}

func (e *setupEvent) record(signalEntityID uuid.UUID) {
	e.signals = append(e.signals, signalEntityID)
}

// withSetupEvent marks the context as ingesting a setup event; signals notified under it are recorded.
func withSetupEvent(ctx context.Context) (context.Context, *setupEvent) {
	recorder := &setupEvent{}
	return context.WithValue(ctx, setupEventKey{}, recorder), recorder
}
