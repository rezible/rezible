package knowledgegraph

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	"github.com/rezible/rezible/ent/schema/schematypes"
)

var stateTestStart = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

func stateTestDay(day int) time.Time {
	return stateTestStart.AddDate(0, 0, day-1)
}

// normalizedRecord is what order independence compares: field values with absent and unset equal.
type normalizedRecord struct {
	displayName, description        any
	properties                      map[string]any
	replacedAt                      *time.Time
	first, last, withdrawn, expires *time.Time
}

func normalizeRecord(record AliasRecord) normalizedRecord {
	value := func(v *schematypes.KnowledgeStateValue) any {
		if v == nil || v.Unset {
			return nil
		}
		return v.Value
	}
	properties := make(map[string]any)
	for key, v := range record.State.Properties {
		if !v.Unset {
			properties[key] = v.Value
		}
	}
	return normalizedRecord{
		displayName: value(record.State.DisplayName),
		description: value(record.State.Description),
		properties:  properties,
		replacedAt:  record.State.ReplacedAt,
		first:       record.FirstObservedAt,
		last:        record.LastObservedAt,
		withdrawn:   record.WithdrawnAt,
		expires:     record.ExpiresAt,
	}
}

func randomEvidence(rng *rand.Rand) AliasEvidence {
	evidence := AliasEvidence{
		Kind:        ke.KindObserved,
		StateMode:   ke.StateModeMerge,
		EffectiveAt: stateTestDay(1 + rng.IntN(4)),
	}
	if rng.IntN(5) == 0 {
		evidence.Kind = ke.KindDeleted
		return evidence
	}
	if rng.IntN(3) == 0 {
		evidence.StateMode = ke.StateModeReplace
	}
	if rng.IntN(2) == 0 {
		expires := evidence.EffectiveAt.Add(time.Duration(1+rng.IntN(3)) * time.Hour)
		evidence.ExpiresAt = &expires
	}
	if rng.IntN(2) == 0 {
		evidence.State.DisplayName = []string{"a", "b", "c"}[rng.IntN(3)]
	}
	evidence.State.Properties = make(map[string]any)
	for _, key := range []string{"tier", "lifecycle", "region"} {
		switch rng.IntN(4) {
		case 0:
			evidence.State.Properties[key] = nil
		case 1:
			evidence.State.Properties[key] = rng.IntN(3)
		}
	}
	return evidence
}

func TestAliasStateOrderIndependence(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for range 2000 {
		evidence := make([]AliasEvidence, 1+rng.IntN(6))
		for i := range evidence {
			evidence[i] = randomEvidence(rng)
		}
		// Every arrival order: a shuffled slice is the ingestion order the rebuild sees.
		for range 6 {
			rng.Shuffle(len(evidence), func(i, j int) { evidence[i], evidence[j] = evidence[j], evidence[i] })
			var incremental AliasRecord
			for _, row := range evidence {
				incremental = ApplyEvidence(incremental, row)
			}
			require.Equal(t, normalizeRecord(RebuildAliasState(evidence)), normalizeRecord(incremental), "%+v", evidence)
		}
	}
}

// TestMergeSubjectStateWorkedExample is the spec's worked example: Grafana lists checkout with no properties,
// a catalogue snapshots it, an alert definition updates a key, and the catalogue snapshots it again without
// lifecycle.
func TestMergeSubjectStateWorkedExample(t *testing.T) {
	grafanaExpiry := stateTestDay(6).Add(3 * time.Hour)
	grafana := ApplyEvidence(AliasRecord{}, AliasEvidence{
		Kind: ke.KindObserved, StateMode: ke.StateModeMerge, EffectiveAt: stateTestDay(6), ExpiresAt: &grafanaExpiry,
		State: schematypes.KnowledgeGraphSubjectState{DisplayName: "checkout"},
	})
	catalogue := RebuildAliasState([]AliasEvidence{
		{Kind: ke.KindObserved, StateMode: ke.StateModeReplace, EffectiveAt: stateTestDay(1),
			State: schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"tier": "core", "lifecycle": "production"}}},
		{Kind: ke.KindObserved, StateMode: ke.StateModeReplace, EffectiveAt: stateTestDay(6),
			State: schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"tier": "core"}}},
	})
	alert := ApplyEvidence(AliasRecord{}, AliasEvidence{
		Kind: ke.KindObserved, StateMode: ke.StateModeMerge, EffectiveAt: stateTestDay(3),
		State: schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"tier": "critical"}},
	})
	assert.True(t, catalogue.State.Properties["lifecycle"].Unset)

	aliases := []SubjectAlias{
		{ID: uuid.New(), CreatedAt: stateTestDay(1), Record: grafana},
		{ID: uuid.New(), CreatedAt: stateTestDay(1), Record: catalogue},
		{ID: uuid.New(), CreatedAt: stateTestDay(3), Record: alert},
	}
	merged := MergeSubjectState(aliases)
	assert.Equal(t, "checkout", merged.State.DisplayName)
	assert.Equal(t, map[string]any{"tier": "core"}, merged.State.Properties)
	assert.Equal(t, aliases[1].ID, merged.Sources.Properties["tier"].AliasID)
	assert.Nil(t, merged.ExpiresAt, "the catalogue alias never expires")

	withoutCatalogue := MergeSubjectState([]SubjectAlias{aliases[0]})
	assert.True(t, IsStale(withoutCatalogue.ExpiresAt, grafanaExpiry.Add(time.Minute)))
}

func TestMergeSubjectStateWithdrawal(t *testing.T) {
	observed := ApplyEvidence(AliasRecord{}, AliasEvidence{Kind: ke.KindObserved, EffectiveAt: stateTestDay(1),
		State: schematypes.KnowledgeGraphSubjectState{DisplayName: "Search"}})
	deleted := ApplyEvidence(observed, AliasEvidence{Kind: ke.KindDeleted, EffectiveAt: stateTestDay(2)})

	withdrawn := MergeSubjectState([]SubjectAlias{{ID: uuid.New(), Record: deleted}})
	assert.Equal(t, new(stateTestDay(2)), withdrawn.WithdrawnAt)
	assert.Equal(t, "Search", withdrawn.State.DisplayName, "a withdrawn subject keeps its last known name")

	stillLive := MergeSubjectState([]SubjectAlias{{ID: uuid.New(), Record: deleted}, {ID: uuid.New(), Record: observed}})
	assert.Nil(t, stillLive.WithdrawnAt)

	assert.Nil(t, MergeSubjectState(nil).WithdrawnAt, "a subject without aliases is never withdrawn")
}
