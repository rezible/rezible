package situations

import (
	"bytes"
	"cmp"
	"slices"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	sitsig "github.com/rezible/rezible/ent/situationsignal"
	"github.com/rezible/rezible/pkg/knowledgegraph"
)

// CorrelationInput is what correlating one signal reads, as of Now.
type CorrelationInput struct {
	Now    time.Time
	Signal Signal
	// EntityIDs are the signal's runtime entities.
	EntityIDs mapset.Set[uuid.UUID]
	// Relationships are the depends-on and adjacent relationships touching the signal's entities.
	Relationships []CorrelationRelationship
	// Situations are the unclosed situations with a matching entity within one relationship of the signal.
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
	// Active is true when any of its signals is active.
	Active bool
	// LastActivityAt is the latest of its signals' attachment and last activity times.
	LastActivityAt    time.Time
	MatchingEntityIDs mapset.Set[uuid.UUID]
}

// activeWithin reports whether the situation has an active signal or was last active within the window.
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

// CorrelationDecision is where a signal is placed. A signal that is not placed has an empty MatchKind.
type CorrelationDecision struct {
	// SituationID is the situation to attach to; uuid.Nil when the signal starts a candidate.
	SituationID uuid.UUID
	// MatchKind is MatchKindSeed when the signal starts a candidate.
	MatchKind sitsig.MatchKind
	// ViaRelationshipID is the lowest-ID relationship that justified a one-hop match.
	ViaRelationshipID *uuid.UUID
}

func (d CorrelationDecision) Placed() bool {
	return d.MatchKind != ""
}

// correlationMatch is a situation that matched one step, with the relationship that justified it.
type correlationMatch struct {
	situation CorrelationSituation
	via       *uuid.UUID
}

// correlationStep matches the situations active within its window whose matching entities meet the
// signal's entities.
type correlationStep struct {
	kind   sitsig.MatchKind
	window time.Duration
	// ranked steps join the best-ranked match; the others join only an unambiguous match.
	ranked bool
	// relates reports whether a relationship relates the signal's entities to a situation's matching
	// entities; nil for the shared entity step.
	relates func(r CorrelationRelationship, signal, matching mapset.Set[uuid.UUID]) bool
}

var correlationSteps = []correlationStep{
	{
		kind:   sitsig.MatchKindSharedEntity,
		window: SharedEntityWindow,
		ranked: true,
	},
	{
		kind:   sitsig.MatchKindDependency,
		window: OneHopWindow,
		ranked: true,
		relates: func(r CorrelationRelationship, signal, matching mapset.Set[uuid.UUID]) bool {
			return r.is(knowledgegraph.RelationshipClassDependsOn) && matching.Contains(r.SourceEntityID) && signal.Contains(r.TargetEntityID)
		},
	},
	{
		kind:   sitsig.MatchKindDependent,
		window: OneHopWindow,
		relates: func(r CorrelationRelationship, signal, matching mapset.Set[uuid.UUID]) bool {
			return r.is(knowledgegraph.RelationshipClassDependsOn) && signal.Contains(r.SourceEntityID) && matching.Contains(r.TargetEntityID)
		},
	},
	{
		kind:   sitsig.MatchKindAdjacent,
		window: OneHopWindow,
		relates: func(r CorrelationRelationship, signal, matching mapset.Set[uuid.UUID]) bool {
			matchDirectionA := signal.Contains(r.SourceEntityID) && matching.Contains(r.TargetEntityID)
			matchDirectionB := signal.Contains(r.TargetEntityID) && matching.Contains(r.SourceEntityID)
			return r.is(knowledgegraph.RelationshipClassAdjacent) && (matchDirectionA || matchDirectionB)
		},
	},
}

func (r CorrelationRelationship) is(class knowledgegraph.RelationshipClass) bool {
	return slices.Contains(class.Predicates(), r.Predicate)
}

// DecideCorrelation places a signal. The first step with a match decides: a shared entity or dependency
// match joins the best-ranked situation; a dependent or adjacent match joins only when it is the only one.
// Otherwise a seeding signal starts a candidate and any other signal is not placed.
func DecideCorrelation(input CorrelationInput) CorrelationDecision {
	relationships := slices.Clone(input.Relationships)
	slices.SortFunc(relationships, func(a, b CorrelationRelationship) int {
		return bytes.Compare(a.ID[:], b.ID[:])
	})
	for _, step := range correlationSteps {
		matches := step.match(input, relationships)
		if len(matches) == 0 {
			continue
		}
		if step.ranked {
			slices.SortFunc(matches, func(a, b correlationMatch) int {
				return cmp.Or(
					cmp.Compare(a.situation.rank(), b.situation.rank()),
					b.situation.LastActivityAt.Compare(a.situation.LastActivityAt),
					bytes.Compare(a.situation.ID[:], b.situation.ID[:]),
				)
			})
		} else if len(matches) > 1 {
			// Several one-hop matches are ambiguous; later steps are not considered.
			break
		}
		return CorrelationDecision{
			SituationID:       matches[0].situation.ID,
			MatchKind:         step.kind,
			ViaRelationshipID: matches[0].via,
		}
	}
	if input.Signal.Seeds() {
		return CorrelationDecision{MatchKind: sitsig.MatchKindSeed}
	}
	return CorrelationDecision{}
}

func (step correlationStep) match(input CorrelationInput, relationships []CorrelationRelationship) []correlationMatch {
	var matches []correlationMatch
	for _, situation := range input.Situations {
		if !situation.activeWithin(input.Now, step.window) {
			continue
		}
		if step.relates == nil {
			if input.EntityIDs.ContainsAnyElement(situation.MatchingEntityIDs) {
				matches = append(matches, correlationMatch{situation: situation})
			}
			continue
		}
		for _, relationship := range relationships {
			if step.relates(relationship, input.EntityIDs, situation.MatchingEntityIDs) {
				matches = append(matches, correlationMatch{situation: situation, via: &relationship.ID})
				break
			}
		}
	}
	return matches
}
