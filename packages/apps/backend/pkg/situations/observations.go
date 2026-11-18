// Package situations holds the pure situation decisions over observations.
package situations

import (
	"bytes"
	"maps"
	"slices"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
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

// RuntimeEntity is one of a situation's runtime entities. It is matching when an observation that is not broad
// touches it.
type RuntimeEntity struct {
	Entity   *ent.KnowledgeEntity
	Matching bool
}

// RuntimeEntities are the distinct entities of the observations' scopes, loaded with
// ent.SituationObservationQuery.WithScopes, ordered by ID.
func RuntimeEntities(observations []*ent.SituationObservation) []RuntimeEntity {
	byID := make(map[uuid.UUID]*RuntimeEntity)
	for _, observation := range observations {
		scope := observation.Scope()
		if scope == nil {
			continue
		}
		for _, row := range scope.Edges.Entities {
			// A merge committed between loading the scope rows and their entities hides the merged entity. The
			// merge moved the row to the survivor, which the next read loads.
			if row.Edges.KnowledgeEntity == nil {
				continue
			}
			entity, seen := byID[row.KnowledgeEntityID]
			if !seen {
				entity = &RuntimeEntity{Entity: row.Edges.KnowledgeEntity}
				byID[row.KnowledgeEntityID] = entity
			}
			entity.Matching = entity.Matching || !scope.Broad
		}
	}
	entities := make([]RuntimeEntity, 0, len(byID))
	for _, id := range slices.SortedFunc(maps.Keys(byID), func(a, b uuid.UUID) int { return bytes.Compare(a[:], b[:]) }) {
		entities = append(entities, *byID[id])
	}
	return entities
}
