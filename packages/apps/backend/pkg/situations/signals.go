// Package situations holds the pure situation decisions over signals.
package situations

import (
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/schema/schematypes"
	ssa "github.com/rezible/rezible/ent/situationsignalattention"
)

// AlertFactsMaxInstances is the most active instances, with labels, described in an alert signal's facts.
const AlertFactsMaxInstances = 5

// Seeds reports whether the signal may start a candidate: its severity is not info and its source's attention is not join-only.
func Seeds(signal rez.SituationSignal) bool {
	return signal.Severity != schematypes.SignalSeverityInfo && signal.SignalAttention != ssa.LevelJoinOnly
}

// EligibleAt reports whether the signal may still be placed: it has not finished more than OneHopWindow
// before now. A signal that resolved but has not finished, such as an episode in its flap grace, is eligible.
func EligibleAt(signal rez.SituationSignal, now time.Time) bool {
	return signal.FinishedAt == nil || now.Sub(*signal.FinishedAt) <= OneHopWindow
}
