package situations

import (
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	rez "github.com/rezible/rezible"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/schema/schematypes"
	ssa "github.com/rezible/rezible/ent/situationsignalattention"
)

var correlationNow = time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

// correlationIDs names entities, situations and relationships so cases can refer to them by name. Names
// sort in the same order as their IDs.
type correlationIDs map[string]uuid.UUID

func (ids correlationIDs) id(name string) uuid.UUID {
	if id, exists := ids[name]; exists {
		return id
	}
	id := uuid.UUID{}
	copy(id[:], name)
	ids[name] = id
	return id
}

func (ids correlationIDs) set(names ...string) mapset.Set[uuid.UUID] {
	set := mapset.NewSet[uuid.UUID]()
	for _, name := range names {
		set.Add(ids.id(name))
	}
	return set
}

type testSituation struct {
	name          string
	raised, muted bool
	active        bool
	// quiet is how long ago the situation was last active.
	quiet    time.Duration
	matching []string
}

type testRelationship struct {
	name, source, predicate, target string
}

func TestDecideCorrelation(t *testing.T) {
	seeding := rez.SituationObservation{Severity: schematypes.SignalSeverityCritical, SignalAttention: ssa.LevelDefault}
	info := rez.SituationObservation{Severity: schematypes.SignalSeverityInfo, SignalAttention: ssa.LevelDefault}
	joinOnly := rez.SituationObservation{Severity: schematypes.SignalSeverityCritical, SignalAttention: ssa.LevelJoinOnly}
	broadEntities := []string{"checkout", "payments", "search", "cart", "fleet"}

	cases := []struct {
		name          string
		observation   rez.SituationObservation
		entities      []string
		relationships []testRelationship
		situations    []testSituation
		// want is the situation joined, "seed" for a new candidate, or "" when not placed.
		want     string
		wantKind MatchKind
	}{
		{
			name:        "shared entity",
			observation: seeding,
			entities:    []string{"checkout"},
			situations:  []testSituation{{name: "s1", active: true, matching: []string{"checkout"}}},
			want:        "s1", wantKind: MatchKindSharedEntity,
		},
		{
			name:        "shared entity within its window of a quiet situation",
			observation: seeding,
			entities:    []string{"checkout"},
			situations:  []testSituation{{name: "s1", quiet: 5 * time.Hour, matching: []string{"checkout"}}},
			want:        "s1", wantKind: MatchKindSharedEntity,
		},
		{
			name:        "shared entity beyond its window starts a candidate",
			observation: seeding,
			entities:    []string{"checkout"},
			situations:  []testSituation{{name: "s1", quiet: 7 * time.Hour, matching: []string{"checkout"}}},
			want:        "seed", wantKind: MatchKindSeed,
		},
		{
			name:        "ranking: raised before candidates before muted",
			observation: seeding,
			entities:    []string{"checkout"},
			situations: []testSituation{
				{name: "s1", muted: true, raised: true, active: true, matching: []string{"checkout"}},
				{name: "s2", active: true, matching: []string{"checkout"}},
				{name: "s3", raised: true, quiet: time.Hour, matching: []string{"checkout"}},
			},
			want: "s3", wantKind: MatchKindSharedEntity,
		},
		{
			name:        "ranking: most recent activity, then lowest ID",
			observation: seeding,
			entities:    []string{"checkout"},
			situations: []testSituation{
				{name: "s1", quiet: 2 * time.Hour, matching: []string{"checkout"}},
				{name: "s3", quiet: time.Hour, matching: []string{"checkout"}},
				{name: "s2", quiet: time.Hour, matching: []string{"checkout"}},
			},
			want: "s2", wantKind: MatchKindSharedEntity,
		},
		{
			name:        "shared entity before dependency",
			observation: seeding,
			entities:    []string{"search"},
			relationships: []testRelationship{
				{name: "r1", source: "checkout", predicate: "calls", target: "search"},
			},
			situations: []testSituation{
				{name: "s1", raised: true, active: true, matching: []string{"checkout"}},
				{name: "s2", active: true, matching: []string{"search"}},
			},
			want: "s2", wantKind: MatchKindSharedEntity,
		},
		{
			name:        "dependency through the lowest-ID relationship",
			observation: seeding,
			entities:    []string{"search", "index"},
			relationships: []testRelationship{
				{name: "r2", source: "checkout", predicate: "calls", target: "search"},
				{name: "r1", source: "checkout", predicate: "reads_from", target: "index"},
			},
			situations: []testSituation{{name: "s1", active: true, matching: []string{"checkout"}}},
			want:       "s1", wantKind: MatchKindDependency,
		},
		{
			name:          "dependency beyond the one-hop window",
			observation:   seeding,
			entities:      []string{"search"},
			relationships: []testRelationship{{name: "r1", source: "checkout", predicate: "calls", target: "search"}},
			situations:    []testSituation{{name: "s1", quiet: time.Hour, matching: []string{"checkout"}}},
			want:          "seed", wantKind: MatchKindSeed,
		},
		{
			name:          "one dependent match",
			observation:   seeding,
			entities:      []string{"checkout"},
			relationships: []testRelationship{{name: "r1", source: "checkout", predicate: "calls", target: "search"}},
			situations:    []testSituation{{name: "s1", quiet: 10 * time.Minute, matching: []string{"search"}}},
			want:          "s1", wantKind: MatchKindDependent,
		},
		{
			name:        "ambiguous dependent matches start a candidate",
			observation: seeding,
			entities:    []string{"checkout"},
			relationships: []testRelationship{
				{name: "r1", source: "checkout", predicate: "calls", target: "search"},
				{name: "r2", source: "checkout", predicate: "writes_to", target: "orders"},
			},
			situations: []testSituation{
				{name: "s1", active: true, matching: []string{"search"}},
				{name: "s2", active: true, matching: []string{"orders"}},
			},
			want: "seed", wantKind: MatchKindSeed,
		},
		{
			name:        "ambiguous adjacent matches leave a non-seeding observation unplaced",
			observation: info,
			entities:    []string{"queue"},
			relationships: []testRelationship{
				{name: "r1", source: "orders", predicate: "publishes_to", target: "queue"},
				{name: "r2", source: "billing", predicate: "consumes_from", target: "queue"},
			},
			situations: []testSituation{
				{name: "s1", active: true, matching: []string{"orders"}},
				{name: "s2", active: true, matching: []string{"billing"}},
			},
			want: "",
		},
		{
			name:          "adjacent in either direction",
			observation:   seeding,
			entities:      []string{"queue"},
			relationships: []testRelationship{{name: "r1", source: "orders", predicate: "publishes_to", target: "queue"}},
			situations:    []testSituation{{name: "s1", active: true, matching: []string{"orders"}}},
			want:          "s1", wantKind: MatchKindAdjacent,
		},
		{
			name:          "structural relationships do not match",
			observation:   seeding,
			entities:      []string{"checkout"},
			relationships: []testRelationship{{name: "r1", source: "shop", predicate: "contains", target: "checkout"}},
			situations:    []testSituation{{name: "s1", active: true, matching: []string{"shop"}}},
			want:          "seed", wantKind: MatchKindSeed,
		},
		{
			name:        "a broad observation joins on a shared entity",
			observation: seeding,
			entities:    broadEntities,
			situations:  []testSituation{{name: "s1", active: true, matching: []string{"fleet"}}},
			want:        "s1", wantKind: MatchKindSharedEntity,
		},
		{
			name:        "a broad observation with no match starts a candidate",
			observation: seeding,
			entities:    broadEntities,
			want:        "seed", wantKind: MatchKindSeed,
		},
		{
			name:        "an info observation joins",
			observation: info,
			entities:    []string{"checkout"},
			situations:  []testSituation{{name: "s1", active: true, matching: []string{"checkout"}}},
			want:        "s1", wantKind: MatchKindSharedEntity,
		},
		{name: "an info observation never starts a candidate", observation: info, entities: []string{"checkout"}, want: ""},
		{name: "a join-only observation never starts a candidate", observation: joinOnly, entities: []string{"checkout"}, want: ""},
		{name: "an observation with no entities starts a candidate", observation: seeding, want: "seed", wantKind: MatchKindSeed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ids := correlationIDs{}
			input := CorrelationInput{Now: correlationNow, Observation: tc.observation, EntityIDs: ids.set(tc.entities...)}
			for _, r := range tc.relationships {
				relationship := CorrelationRelationship{
					ID:             ids.id(r.name),
					Predicate:      knr.Predicate(r.predicate),
					SourceEntityID: ids.id(r.source),
					TargetEntityID: ids.id(r.target),
				}
				input.Relationships = append(input.Relationships, relationship)
			}
			for _, s := range tc.situations {
				situation := CorrelationSituation{
					ID:                ids.id(s.name),
					Raised:            s.raised,
					Muted:             s.muted,
					Active:            s.active,
					LastActivityAt:    correlationNow.Add(-s.quiet),
					MatchingEntityIDs: ids.set(s.matching...),
				}
				input.Situations = append(input.Situations, situation)
			}

			decision := DecideCorrelation(input)
			assert.Equal(t, tc.want != "", decision.Placed())
			assert.Equal(t, tc.wantKind, decision.MatchKind)
			wantSituation := uuid.Nil
			if tc.want != "" && tc.want != "seed" {
				wantSituation = ids.id(tc.want)
			}
			assert.Equal(t, wantSituation, decision.SituationID)
		})
	}
}
