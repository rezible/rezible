package situations

import (
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sitsig "github.com/rezible/rezible/ent/situationsignal"
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
	seeding := Signal{Severity: schematypes.SignalSeverityCritical, SignalAttention: ssa.LevelDefault}
	info := Signal{Severity: schematypes.SignalSeverityInfo, SignalAttention: ssa.LevelDefault}
	joinOnly := Signal{Severity: schematypes.SignalSeverityCritical, SignalAttention: ssa.LevelJoinOnly}
	broadEntities := []string{"checkout", "payments", "search", "cart", "fleet"}

	cases := []struct {
		name          string
		signal        Signal
		entities      []string
		relationships []testRelationship
		situations    []testSituation
		// want is the situation joined, "seed" for a new candidate, or "" when not placed.
		want     string
		wantKind sitsig.MatchKind
		wantVia  string
	}{
		{
			name:       "shared entity",
			signal:     seeding,
			entities:   []string{"checkout"},
			situations: []testSituation{{name: "s1", active: true, matching: []string{"checkout"}}},
			want:       "s1", wantKind: sitsig.MatchKindSharedEntity,
		},
		{
			name:       "shared entity within its window of a quiet situation",
			signal:     seeding,
			entities:   []string{"checkout"},
			situations: []testSituation{{name: "s1", quiet: 5 * time.Hour, matching: []string{"checkout"}}},
			want:       "s1", wantKind: sitsig.MatchKindSharedEntity,
		},
		{
			name:       "shared entity beyond its window starts a candidate",
			signal:     seeding,
			entities:   []string{"checkout"},
			situations: []testSituation{{name: "s1", quiet: 7 * time.Hour, matching: []string{"checkout"}}},
			want:       "seed", wantKind: sitsig.MatchKindSeed,
		},
		{
			name:     "ranking: raised before candidates before muted",
			signal:   seeding,
			entities: []string{"checkout"},
			situations: []testSituation{
				{name: "s1", muted: true, raised: true, active: true, matching: []string{"checkout"}},
				{name: "s2", active: true, matching: []string{"checkout"}},
				{name: "s3", raised: true, quiet: time.Hour, matching: []string{"checkout"}},
			},
			want: "s3", wantKind: sitsig.MatchKindSharedEntity,
		},
		{
			name:     "ranking: most recent activity, then lowest ID",
			signal:   seeding,
			entities: []string{"checkout"},
			situations: []testSituation{
				{name: "s1", quiet: 2 * time.Hour, matching: []string{"checkout"}},
				{name: "s3", quiet: time.Hour, matching: []string{"checkout"}},
				{name: "s2", quiet: time.Hour, matching: []string{"checkout"}},
			},
			want: "s2", wantKind: sitsig.MatchKindSharedEntity,
		},
		{
			name:     "shared entity before dependency",
			signal:   seeding,
			entities: []string{"search"},
			relationships: []testRelationship{
				{name: "r1", source: "checkout", predicate: "calls", target: "search"},
			},
			situations: []testSituation{
				{name: "s1", raised: true, active: true, matching: []string{"checkout"}},
				{name: "s2", active: true, matching: []string{"search"}},
			},
			want: "s2", wantKind: sitsig.MatchKindSharedEntity,
		},
		{
			name:     "dependency through the lowest-ID relationship",
			signal:   seeding,
			entities: []string{"search", "index"},
			relationships: []testRelationship{
				{name: "r2", source: "checkout", predicate: "calls", target: "search"},
				{name: "r1", source: "checkout", predicate: "reads_from", target: "index"},
			},
			situations: []testSituation{{name: "s1", active: true, matching: []string{"checkout"}}},
			want:       "s1", wantKind: sitsig.MatchKindDependency, wantVia: "r1",
		},
		{
			name:          "dependency beyond the one-hop window",
			signal:        seeding,
			entities:      []string{"search"},
			relationships: []testRelationship{{name: "r1", source: "checkout", predicate: "calls", target: "search"}},
			situations:    []testSituation{{name: "s1", quiet: time.Hour, matching: []string{"checkout"}}},
			want:          "seed", wantKind: sitsig.MatchKindSeed,
		},
		{
			name:          "one dependent match",
			signal:        seeding,
			entities:      []string{"checkout"},
			relationships: []testRelationship{{name: "r1", source: "checkout", predicate: "calls", target: "search"}},
			situations:    []testSituation{{name: "s1", quiet: 10 * time.Minute, matching: []string{"search"}}},
			want:          "s1", wantKind: sitsig.MatchKindDependent, wantVia: "r1",
		},
		{
			name:     "ambiguous dependent matches start a candidate",
			signal:   seeding,
			entities: []string{"checkout"},
			relationships: []testRelationship{
				{name: "r1", source: "checkout", predicate: "calls", target: "search"},
				{name: "r2", source: "checkout", predicate: "writes_to", target: "orders"},
			},
			situations: []testSituation{
				{name: "s1", active: true, matching: []string{"search"}},
				{name: "s2", active: true, matching: []string{"orders"}},
			},
			want: "seed", wantKind: sitsig.MatchKindSeed,
		},
		{
			name:     "ambiguous adjacent matches leave a non-seeding signal unplaced",
			signal:   info,
			entities: []string{"queue"},
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
			signal:        seeding,
			entities:      []string{"queue"},
			relationships: []testRelationship{{name: "r1", source: "orders", predicate: "publishes_to", target: "queue"}},
			situations:    []testSituation{{name: "s1", active: true, matching: []string{"orders"}}},
			want:          "s1", wantKind: sitsig.MatchKindAdjacent, wantVia: "r1",
		},
		{
			name:          "structural relationships do not match",
			signal:        seeding,
			entities:      []string{"checkout"},
			relationships: []testRelationship{{name: "r1", source: "shop", predicate: "contains", target: "checkout"}},
			situations:    []testSituation{{name: "s1", active: true, matching: []string{"shop"}}},
			want:          "seed", wantKind: sitsig.MatchKindSeed,
		},
		{
			name:       "a broad signal joins on a shared entity",
			signal:     seeding,
			entities:   broadEntities,
			situations: []testSituation{{name: "s1", active: true, matching: []string{"fleet"}}},
			want:       "s1", wantKind: sitsig.MatchKindSharedEntity,
		},
		{
			name:     "a broad signal with no match starts a candidate",
			signal:   seeding,
			entities: broadEntities,
			want:     "seed", wantKind: sitsig.MatchKindSeed,
		},
		{
			name:       "an info signal joins",
			signal:     info,
			entities:   []string{"checkout"},
			situations: []testSituation{{name: "s1", active: true, matching: []string{"checkout"}}},
			want:       "s1", wantKind: sitsig.MatchKindSharedEntity,
		},
		{name: "an info signal never starts a candidate", signal: info, entities: []string{"checkout"}, want: ""},
		{name: "a join-only signal never starts a candidate", signal: joinOnly, entities: []string{"checkout"}, want: ""},
		{name: "a signal with no entities starts a candidate", signal: seeding, want: "seed", wantKind: sitsig.MatchKindSeed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ids := correlationIDs{}
			input := CorrelationInput{Now: correlationNow, Signal: tc.signal, EntityIDs: ids.set(tc.entities...)}
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
			if tc.wantVia == "" {
				assert.Nil(t, decision.ViaRelationshipID)
			} else if assert.NotNil(t, decision.ViaRelationshipID) {
				assert.Equal(t, ids.id(tc.wantVia), *decision.ViaRelationshipID)
			}
		})
	}
}
