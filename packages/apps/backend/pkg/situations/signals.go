// Package situations holds the signal contract and the pure situation decisions over it.
package situations

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/rezible/rezible/ent/schema/schematypes"
	ssa "github.com/rezible/rezible/ent/situationsignalattention"
)

// AlertFactsMaxInstances is the most active instances, with labels, described in an alert signal's facts.
const AlertFactsMaxInstances = 5

// SignalKind names the kind of thing a signal is.
type SignalKind string

const SignalKindAlertEpisode SignalKind = "alert_episode"

// Signal is something that can be evidence in a situation, filled by its source from stored state.
type Signal struct {
	// EntityID is the signal's knowledge entity.
	EntityID uuid.UUID
	Kind     SignalKind
	Title    string
	// SourceEntityID is the knowledge entity of what produced the signal, such as an alert definition; nil when none.
	SourceEntityID  *uuid.UUID
	SignalAttention ssa.Level
	// SignalAttentionSetAt is when a person last set the source's attention; nil when never set.
	SignalAttentionSetAt *time.Time
	// Severity is the signal's highest severity.
	Severity  schematypes.SignalSeverity
	StartedAt time.Time
	// FinishedAt is nil while the signal is unfinished.
	FinishedAt *time.Time
	// Active is true while the signal is currently firing.
	Active bool
	// LastActivityAt is when Rezible last received evidence of the signal.
	LastActivityAt time.Time
	EventIDs       []uuid.UUID
	// Revision is the number of linked events; it grows as evidence arrives.
	Revision int
	// Alert is set for SignalKindAlertEpisode.
	Alert *schematypes.SituationAlertFacts
}

// SituationSignalSource loads the signals of one kind from its stored state.
type SituationSignalSource interface {
	SituationSignalKind() SignalKind
	LoadSignals(ctx context.Context, entityIDs []uuid.UUID, opts LoadSignalsOptions) ([]Signal, error)
}

// LoadSignalsOptions asks for facts that cost more than stored state, which only evaluation needs.
type LoadSignalsOptions struct {
	// History adds facts computed from the source's history, such as an alert's baseline.
	History bool
}

// Seeds reports whether the signal may start a candidate: its severity is not info and its source's
// attention is not join-only.
func (s Signal) Seeds() bool {
	return s.Severity != schematypes.SignalSeverityInfo && s.SignalAttention != ssa.LevelJoinOnly
}

// EligibleAt reports whether the signal may still be placed: it has not finished more than OneHopWindow
// before now. A signal that resolved but has not finished, such as an episode in its flap grace, is eligible.
func (s Signal) EligibleAt(now time.Time) bool {
	return s.FinishedAt == nil || now.Sub(*s.FinishedAt) <= OneHopWindow
}
