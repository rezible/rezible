// Package situations holds the pure situation decisions over observations.
package situations

import (
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/schema/schematypes"
	ssa "github.com/rezible/rezible/ent/situationsignalattention"
)

// AlertFactsMaxInstances is the most active instances, with labels, described in an alert observation's facts.
const AlertFactsMaxInstances = 5

// Seeds reports whether the observation may start a candidate: its severity is not info and its signal's
// attention is not join-only.
func Seeds(observation rez.SituationObservation) bool {
	return observation.Severity != schematypes.SignalSeverityInfo && observation.SignalAttention != ssa.LevelJoinOnly
}

// EligibleAt reports whether the observation may still be placed: it has not finished more than OneHopWindow
// before now. An observation that resolved but has not finished, such as an episode in its flap grace, is
// eligible.
func EligibleAt(observation rez.SituationObservation, now time.Time) bool {
	return observation.FinishedAt == nil || now.Sub(*observation.FinishedAt) <= OneHopWindow
}
