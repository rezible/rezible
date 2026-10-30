package situations

import (
	"bytes"
	"cmp"
	"slices"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/pkg/knowledgegraph"
)

// CorrelationInput is what correlating one observation reads, as of Now.
type CorrelationInput struct {
	Now         time.Time
	Observation rez.SituationObservation
	// EntityIDs are the observation's runtime entities.
	EntityIDs mapset.Set[uuid.UUID]
	// Relationships are the depends-on and adjacent relationships touching the observation's entities.
	Relationships []CorrelationRelationship
	// Situations are the unclosed situations with a matching entity within one relationship of the observation.
	Situations []CorrelationSituation
}

type CorrelationRelationship struct {
	ID             uuid.UUID
	Predicate      knr.Predicate
	SourceEntityID uuid.UUID
	TargetEntityID uuid.UUID
}

type CorrelationSituation struct {
	ID     uuid.UUID
	Raised bool
	Muted  bool
	// Active is true when any of its observations is active.
	Active bool
	// LastActivityAt is the latest of its observations' attachment and last activity times.
	LastActivityAt    time.Time
	MatchingEntityIDs mapset.Set[uuid.UUID]
}

// activeWithin reports whether the situation has an active observation or was last active within the window.
func (s CorrelationSituation) activeWithin(now time.Time, window time.Duration) bool {
	return s.Active || now.Sub(s.LastActivityAt) <= window
}

// rank orders raised and unmuted situations first, then unmuted candidates, then muted ones.
func (s CorrelationSituation) rank() int {
	switch {
	case s.Muted:
		return 2
	case s.Raised:
		return 0
	default:
		return 1
	}
}

// MatchKind is how correlation placed an observation.
type MatchKind string

const (
	MatchKindSeed         MatchKind = "seed"
	MatchKindSharedEntity MatchKind = "shared_entity"
	MatchKindDependency   MatchKind = "dependency"
	MatchKindDependent    MatchKind = "dependent"
	MatchKindAdjacent     MatchKind = "adjacent"
)

// CorrelationDecision is where an observation is placed. An observation that is not placed has an empty
// MatchKind.
type CorrelationDecision struct {
	// SituationID is the situation to attach to; uuid.Nil when the observation starts a candidate.
	SituationID uuid.UUID
	// MatchKind is MatchKindSeed when the observation starts a candidate.
	MatchKind MatchKind
}

func (d CorrelationDecision) Placed() bool {
	return d.MatchKind != ""
}

// correlationStep matches the situations active within its window whose matching entities meet the
// observation's entities.
type correlationStep struct {
	kind   MatchKind
	window time.Duration
	// ranked steps join the best-ranked match; the others join only an unambiguous match.
	ranked bool
	// relates reports whether a relationship relates the observation's entities to a situation's matching
	// entities; nil for the shared entity step.
	relates func(r CorrelationRelationship, observation, matching mapset.Set[uuid.UUID]) bool
}

var correlationSteps = []correlationStep{
	{
		kind:   MatchKindSharedEntity,
		window: SharedEntityWindow,
		ranked: true,
	},
	{
		kind:   MatchKindDependency,
		window: OneHopWindow,
		ranked: true,
		relates: func(r CorrelationRelationship, observation, matching mapset.Set[uuid.UUID]) bool {
			return r.is(knowledgegraph.RelationshipClassDependsOn) && matching.Contains(r.SourceEntityID) && observation.Contains(r.TargetEntityID)
		},
	},
	{
		kind:   MatchKindDependent,
		window: OneHopWindow,
		relates: func(r CorrelationRelationship, observation, matching mapset.Set[uuid.UUID]) bool {
			return r.is(knowledgegraph.RelationshipClassDependsOn) && observation.Contains(r.SourceEntityID) && matching.Contains(r.TargetEntityID)
		},
	},
	{
		kind:   MatchKindAdjacent,
		window: OneHopWindow,
		relates: func(r CorrelationRelationship, observation, matching mapset.Set[uuid.UUID]) bool {
			matchDirectionA := observation.Contains(r.SourceEntityID) && matching.Contains(r.TargetEntityID)
			matchDirectionB := observation.Contains(r.TargetEntityID) && matching.Contains(r.SourceEntityID)
			return r.is(knowledgegraph.RelationshipClassAdjacent) && (matchDirectionA || matchDirectionB)
		},
	},
}

func (r CorrelationRelationship) is(class knowledgegraph.RelationshipClass) bool {
	return slices.Contains(class.Predicates(), r.Predicate)
}

// DecideCorrelation places an observation. The first step with a match decides: a shared entity or dependency
// match joins the best-ranked situation; a dependent or adjacent match joins only when it is the only one.
// Otherwise a seeding observation starts a candidate and any other observation is not placed.
func DecideCorrelation(input CorrelationInput) CorrelationDecision {
	for _, step := range correlationSteps {
		matches := step.match(input)
		if len(matches) == 0 {
			continue
		}
		if step.ranked {
			slices.SortFunc(matches, func(a, b CorrelationSituation) int {
				return cmp.Or(
					cmp.Compare(a.rank(), b.rank()),
					b.LastActivityAt.Compare(a.LastActivityAt),
					bytes.Compare(a.ID[:], b.ID[:]),
				)
			})
		} else if len(matches) > 1 {
			// Several one-hop matches are ambiguous; later steps are not considered.
			break
		}
		return CorrelationDecision{SituationID: matches[0].ID, MatchKind: step.kind}
	}
	if Seeds(input.Observation) {
		return CorrelationDecision{MatchKind: MatchKindSeed}
	}
	return CorrelationDecision{}
}

func (step correlationStep) match(input CorrelationInput) []CorrelationSituation {
	var matches []CorrelationSituation
	for _, situation := range input.Situations {
		if !situation.activeWithin(input.Now, step.window) {
			continue
		}
		if step.relates == nil {
			if input.EntityIDs.ContainsAnyElement(situation.MatchingEntityIDs) {
				matches = append(matches, situation)
			}
			continue
		}
		for _, relationship := range input.Relationships {
			if step.relates(relationship, input.EntityIDs, situation.MatchingEntityIDs) {
				matches = append(matches, situation)
				break
			}
		}
	}
	return matches
}
