package knowledgegraph

import (
	"encoding/binary"
	"math/rand/v2"
	"slices"
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

func randomEvidence(rng *rand.Rand) AliasEvidence {
	evidence := AliasEvidence{
		EventID:     randomEventID(rng),
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

func randomEventID(rng *rand.Rand) uuid.UUID {
	var id uuid.UUID
	binary.BigEndian.PutUint64(id[:8], rng.Uint64())
	binary.BigEndian.PutUint64(id[8:], rng.Uint64())
	return id
}

func applyInOrder(evidence ...AliasEvidence) AliasRecord {
	var record AliasRecord
	for _, row := range evidence {
		record = ApplyEvidence(record, row)
	}
	return record
}

func TestAliasStateOrderIndependence(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for range 2000 {
		evidence := make([]AliasEvidence, 1+rng.IntN(6))
		for i := range evidence {
			evidence[i] = randomEvidence(rng)
			if evidence[i].Kind == ke.KindObserved && rng.IntN(2) == 0 {
				evidence[i].ReobservedAt = new(evidence[i].EffectiveAt.Add(time.Duration(1+rng.IntN(3)) * 24 * time.Hour))
			}
		}
		rebuilt := RebuildAliasState(evidence)
		stored := slices.Clone(evidence)
		for i := range stored {
			row := &stored[i]
			if row.ReobservedAt != nil {
				evidence = append(evidence, AliasEvidence{Kind: ke.KindObserved, EffectiveAt: *row.ReobservedAt, ReobservationOf: row})
				evidence = append(evidence, AliasEvidence{Kind: ke.KindObserved, EffectiveAt: row.ReobservedAt.Add(-time.Hour), ReobservationOf: row})
			}
		}
		for range 6 {
			rng.Shuffle(len(evidence), func(i, j int) { evidence[i], evidence[j] = evidence[j], evidence[i] })
			require.Equal(t, rebuilt, applyInOrder(evidence...), "%+v", evidence)
		}
	}
}

// TestAliasStateTies applies two rows with one effective time in both orders: the later event ID decides.
func TestAliasStateTies(t *testing.T) {
	earlier, later := uuid.MustParse("01900000-0000-7000-8000-000000000001"), uuid.MustParse("01900000-0000-7000-8000-000000000002")
	at := stateTestDay(2)
	observed := func(id uuid.UUID, mode ke.StateMode, state schematypes.KnowledgeGraphSubjectState) AliasEvidence {
		return AliasEvidence{EventID: id, Kind: ke.KindObserved, StateMode: mode, EffectiveAt: at, State: state}
	}
	bothOrders := func(t *testing.T, a, b AliasEvidence) AliasRecord {
		t.Helper()
		record := applyInOrder(a, b)
		require.Equal(t, record, applyInOrder(b, a))
		return record
	}

	t.Run("observe and delete", func(t *testing.T) {
		deleted := AliasEvidence{EventID: later, Kind: ke.KindDeleted, EffectiveAt: at}
		record := bothOrders(t, observed(earlier, ke.StateModeMerge, schematypes.KnowledgeGraphSubjectState{}), deleted)
		assert.Equal(t, &deleted.EffectiveAt, record.WithdrawnAt)

		deleted.EventID = earlier
		record = bothOrders(t, observed(later, ke.StateModeMerge, schematypes.KnowledgeGraphSubjectState{}), deleted)
		assert.Nil(t, record.WithdrawnAt)
	})

	t.Run("two values of one field", func(t *testing.T) {
		record := bothOrders(t,
			observed(earlier, ke.StateModeMerge, schematypes.KnowledgeGraphSubjectState{DisplayName: "old", Properties: map[string]any{"tier": 1}}),
			observed(later, ke.StateModeMerge, schematypes.KnowledgeGraphSubjectState{DisplayName: "new", Properties: map[string]any{"tier": 2}}))
		assert.Equal(t, "new", record.State.DisplayName.Value)
		assert.Equal(t, 2, record.State.Properties["tier"].Value)
	})

	t.Run("replace and merge", func(t *testing.T) {
		record := bothOrders(t,
			observed(earlier, ke.StateModeMerge, schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"tier": 1}}),
			observed(later, ke.StateModeReplace, schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"region": "eu"}}))
		assert.True(t, record.State.Properties["tier"].Unset)

		record = bothOrders(t,
			observed(later, ke.StateModeMerge, schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"tier": 1}}),
			observed(earlier, ke.StateModeReplace, schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"region": "eu"}}))
		assert.Equal(t, 1, record.State.Properties["tier"].Value)
	})
}

// TestAliasStateReplaceGate checks that a snapshot omitting a property unsets it even when the statement it
// supersedes arrives after it.
func TestAliasStateReplaceGate(t *testing.T) {
	stated := AliasEvidence{EventID: uuid.New(), Kind: ke.KindObserved, StateMode: ke.StateModeMerge, EffectiveAt: stateTestDay(1),
		State: schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"p": "value"}}}
	snapshot := AliasEvidence{EventID: uuid.New(), Kind: ke.KindObserved, StateMode: ke.StateModeReplace, EffectiveAt: stateTestDay(2),
		State: schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"q": "value"}}}
	unset := schematypes.KnowledgeStateValue{Unset: true, KnowledgeEvidenceKey: snapshot.Key()}
	assert.Equal(t, unset, applyInOrder(stated, snapshot).State.Properties["p"])
	assert.Equal(t, unset, applyInOrder(snapshot, stated).State.Properties["p"])
}

// TestMergeSubjectStateWorkedExample is a worked example: Grafana lists checkout with no properties,
// a catalogue snapshots it, an alert definition updates a key, and the catalogue snapshots it again without
// lifecycle.
func TestMergeSubjectStateWorkedExample(t *testing.T) {
	grafanaExpiry := stateTestDay(6).Add(3 * time.Hour)
	grafana := ApplyEvidence(AliasRecord{}, AliasEvidence{
		EventID: uuid.New(), Kind: ke.KindObserved, StateMode: ke.StateModeMerge, EffectiveAt: stateTestDay(6), ExpiresAt: &grafanaExpiry,
		State: schematypes.KnowledgeGraphSubjectState{DisplayName: "checkout"},
	})
	catalogue := RebuildAliasState([]AliasEvidence{
		{EventID: uuid.New(), Kind: ke.KindObserved, StateMode: ke.StateModeReplace, EffectiveAt: stateTestDay(1),
			State: schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"tier": "core", "lifecycle": "production"}}},
		{EventID: uuid.New(), Kind: ke.KindObserved, StateMode: ke.StateModeReplace, EffectiveAt: stateTestDay(6),
			State: schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"tier": "core"}}},
	})
	alert := ApplyEvidence(AliasRecord{}, AliasEvidence{
		EventID: uuid.New(), Kind: ke.KindObserved, StateMode: ke.StateModeMerge, EffectiveAt: stateTestDay(3),
		State: schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"tier": "critical"}},
	})
	assert.True(t, catalogue.State.Properties["lifecycle"].Unset)

	aliases := []SubjectAlias{
		{ID: uuid.New(), Authority: schematypes.KnowledgeAuthorityObserved, Record: grafana},
		{ID: uuid.New(), Authority: schematypes.KnowledgeAuthorityDeclared, Record: catalogue},
		{ID: uuid.New(), Authority: schematypes.KnowledgeAuthorityObserved, Record: alert},
	}
	merged := MergeSubjectState(aliases)
	assert.Equal(t, "checkout", merged.State.DisplayName)
	assert.Equal(t, map[string]any{"tier": "core"}, merged.State.Properties)
	assert.Equal(t, aliases[1].ID, merged.Sources.Properties["tier"].AliasID)
	assert.Nil(t, merged.ExpiresAt, "the catalogue alias never expires")

	withoutCatalogue := MergeSubjectState([]SubjectAlias{aliases[0]})
	assert.True(t, IsStale(withoutCatalogue.ExpiresAt, grafanaExpiry.Add(time.Minute)))
}

func TestMergeSubjectStateAuthority(t *testing.T) {
	declared, observed := schematypes.KnowledgeAuthorityDeclared, schematypes.KnowledgeAuthorityObserved
	at := func(day int, eventID uuid.UUID) schematypes.KnowledgeEvidenceKey {
		return schematypes.KnowledgeEvidenceKey{EffectiveAt: stateTestDay(day), EventID: eventID}
	}
	value := func(v any, k schematypes.KnowledgeEvidenceKey) *schematypes.KnowledgeStateValue {
		return &schematypes.KnowledgeStateValue{Value: v, KnowledgeEvidenceKey: k}
	}
	unset := func(k schematypes.KnowledgeEvidenceKey) schematypes.KnowledgeStateValue {
		return schematypes.KnowledgeStateValue{Unset: true, KnowledgeEvidenceKey: k}
	}
	alias := func(id string, authority schematypes.KnowledgeAuthority, state schematypes.KnowledgeAliasState) SubjectAlias {
		return SubjectAlias{ID: uuid.MustParse(id), Authority: authority, Record: AliasRecord{State: state}}
	}
	lowID, highID := "00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002"
	earlyEvent, lateEvent := uuid.MustParse("00000000-0000-0000-0000-00000000000a"), uuid.MustParse("00000000-0000-0000-0000-00000000000b")

	catalogue := alias(lowID, declared, schematypes.KnowledgeAliasState{
		DisplayName: value("Checkout API", at(1, earlyEvent)),
		Properties: map[string]schematypes.KnowledgeStateValue{
			"tier":      *value("core", at(1, earlyEvent)),
			"lifecycle": unset(at(2, earlyEvent)),
		},
	})
	alert := alias(highID, observed, schematypes.KnowledgeAliasState{
		DisplayName: value("checkout", at(5, lateEvent)),
		Description: value("Checkout service", at(5, lateEvent)),
		Properties: map[string]schematypes.KnowledgeStateValue{
			"tier":      *value("critical", at(5, lateEvent)),
			"lifecycle": *value("production", at(5, lateEvent)),
			"runbook":   *value("https://runbooks.test/checkout", at(5, lateEvent)),
		},
	})
	merged := MergeSubjectState([]SubjectAlias{alert, catalogue})
	assert.Equal(t, "Checkout API", merged.State.DisplayName, "a declared name survives a later observed one")
	assert.Equal(t, declared, merged.Sources.DisplayName.Authority)
	assert.Equal(t, "Checkout service", merged.State.Description, "an observed value fills a field no declared alias states")
	assert.Equal(t, map[string]any{"tier": "core", "runbook": "https://runbooks.test/checkout"}, merged.State.Properties,
		"the declared unset hides the observed lifecycle; the runbook it never stated stays")
	assert.Equal(t, catalogue.ID, merged.Sources.Properties["lifecycle"].AliasID, "the hidden field's source is the unsetting alias")
	assert.True(t, stateTestDay(5).Equal(*merged.StateEffectiveAt))

	// Two aliases state the name at the same time: the later event wins, whichever alias came first.
	for _, ids := range [][2]string{{lowID, highID}, {highID, lowID}} {
		first := alias(ids[0], declared, schematypes.KnowledgeAliasState{DisplayName: value("first", at(3, lateEvent))})
		second := alias(ids[1], declared, schematypes.KnowledgeAliasState{DisplayName: value("second", at(3, earlyEvent))})
		for _, aliases := range [][]SubjectAlias{{first, second}, {second, first}} {
			assert.Equal(t, "first", MergeSubjectState(aliases).State.DisplayName)
		}
	}
	// Within one tier an unset does not hide another alias's value.
	sameTier := MergeSubjectState([]SubjectAlias{
		alias(lowID, declared, schematypes.KnowledgeAliasState{Properties: map[string]schematypes.KnowledgeStateValue{"tier": unset(at(4, lateEvent))}}),
		alias(highID, declared, schematypes.KnowledgeAliasState{Properties: map[string]schematypes.KnowledgeStateValue{"tier": *value("core", at(1, earlyEvent))}}),
	})
	assert.Equal(t, map[string]any{"tier": "core"}, sameTier.State.Properties)
}

func TestMergeSubjectStateWithdrawal(t *testing.T) {
	observed := ApplyEvidence(AliasRecord{}, AliasEvidence{EventID: uuid.New(), Kind: ke.KindObserved, EffectiveAt: stateTestDay(1),
		State: schematypes.KnowledgeGraphSubjectState{DisplayName: "Search"}})
	deleted := ApplyEvidence(observed, AliasEvidence{EventID: uuid.New(), Kind: ke.KindDeleted, EffectiveAt: stateTestDay(2)})

	withdrawn := MergeSubjectState([]SubjectAlias{{ID: uuid.New(), Record: deleted}})
	assert.Equal(t, new(stateTestDay(2)), withdrawn.WithdrawnAt)
	assert.Equal(t, "Search", withdrawn.State.DisplayName, "a withdrawn subject keeps its last known name")

	stillLive := MergeSubjectState([]SubjectAlias{{ID: uuid.New(), Record: deleted}, {ID: uuid.New(), Record: observed}})
	assert.Nil(t, stillLive.WithdrawnAt)

	assert.Nil(t, MergeSubjectState(nil).WithdrawnAt, "a subject without aliases is never withdrawn")
}
