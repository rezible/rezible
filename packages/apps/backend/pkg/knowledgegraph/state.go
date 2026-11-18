package knowledgegraph

import (
	"bytes"
	"cmp"
	"encoding/json"
	"slices"
	"time"

	"github.com/google/uuid"

	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	"github.com/rezible/rezible/ent/schema/schematypes"
)

// AliasEvidence is one evidence row as alias state reads it.
type AliasEvidence struct {
	EventID     uuid.UUID
	Kind        ke.Kind
	StateMode   ke.StateMode
	EffectiveAt time.Time
	ExpiresAt   *time.Time
	State       schematypes.KnowledgeGraphSubjectState
	// ReobservedAt is the stored latest repeat, expanded by RebuildAliasState.
	ReobservedAt *time.Time
	// ReobservationOf names the original time-bounded observation; the repeat carries no state.
	ReobservationOf *AliasEvidence
}

// Key is the row's place in its alias's evidence order.
func (e AliasEvidence) Key() schematypes.KnowledgeEvidenceKey {
	if e.ReobservationOf != nil {
		e.EventID = e.ReobservationOf.EventID
	}
	return schematypes.KnowledgeEvidenceKey{EffectiveAt: e.EffectiveAt, EventID: e.EventID}
}

// AliasRecord is an alias's state and observation times. LastObservedAt and WithdrawnAt are the effective
// times of State.LastObserved and State.Withdrawn.
type AliasRecord struct {
	State           schematypes.KnowledgeAliasState
	FirstObservedAt *time.Time
	LastObservedAt  *time.Time
	WithdrawnAt     *time.Time
	ExpiresAt       *time.Time
}

// ApplyEvidence returns the alias record after one more evidence row. Rows are ordered by key, so applying
// them in any order gives the result of RebuildAliasState over them. A row whose key is already stored is a
// replay and changes nothing.
func ApplyEvidence(record AliasRecord, evidence AliasEvidence) AliasRecord {
	if original := evidence.ReobservationOf; original != nil {
		if original.ExpiresAt == nil {
			return record
		}
		evidence.Kind, evidence.StateMode = ke.KindObserved, ke.StateModeMerge
		evidence.State = schematypes.KnowledgeGraphSubjectState{}
		evidence.ExpiresAt = new(evidence.EffectiveAt.Add(original.ExpiresAt.Sub(original.EffectiveAt)))
	}
	k := evidence.Key()
	state := &record.State
	if evidence.Kind == ke.KindDeleted {
		if keyBefore(state.LastObserved, k) {
			state.Withdrawn = latestKey(state.Withdrawn, &k)
		}
	} else {
		*state = applyStatements(*state, evidence)
		if keyBefore(state.LastObserved, k) {
			state.LastObserved = &k
			record.ExpiresAt = evidence.ExpiresAt
		}
		if state.Withdrawn != nil && state.Withdrawn.Compare(k) < 0 {
			state.Withdrawn = nil
		}
		if record.FirstObservedAt == nil || k.EffectiveAt.Before(*record.FirstObservedAt) {
			record.FirstObservedAt = &k.EffectiveAt
		}
	}
	record.LastObservedAt = keyTime(state.LastObserved)
	record.WithdrawnAt = keyTime(state.Withdrawn)
	return record
}

func applyStatements(state schematypes.KnowledgeAliasState, evidence AliasEvidence) schematypes.KnowledgeAliasState {
	k := evidence.Key()
	writable := func(stored *schematypes.KnowledgeStateValue) bool {
		return stored == nil || stored.Compare(k) < 0
	}
	if name := evidence.State.DisplayName; name != "" && writable(state.DisplayName) {
		state.DisplayName = &schematypes.KnowledgeStateValue{Value: name, KnowledgeEvidenceKey: k}
	}
	if description := evidence.State.Description; description != "" && writable(state.Description) {
		state.Description = &schematypes.KnowledgeStateValue{Value: description, KnowledgeEvidenceKey: k}
	}
	properties := make(map[string]schematypes.KnowledgeStateValue, len(state.Properties)+len(evidence.State.Properties))
	for key, stored := range state.Properties {
		properties[key] = stored
	}
	// Only properties are gated by the latest replace: a replace unsets the properties it omits but never the
	// display name or description, so gating those too would depend on arrival order.
	if state.Replaced != nil && k.Compare(*state.Replaced) < 0 {
		// The later replace omitted any property this row states that it has no entry for, and would have
		// unset it had this row arrived first. Every stored entry is at or after the replace, so it stays.
		for key := range evidence.State.Properties {
			if _, found := properties[key]; !found {
				properties[key] = schematypes.KnowledgeStateValue{Unset: true, KnowledgeEvidenceKey: *state.Replaced}
			}
		}
		state.Properties = properties
		return state
	}
	for key, stored := range properties {
		_, stated := evidence.State.Properties[key]
		if evidence.StateMode == ke.StateModeReplace && !stated && writable(&stored) {
			properties[key] = schematypes.KnowledgeStateValue{Unset: true, KnowledgeEvidenceKey: k}
		}
	}
	for key, value := range evidence.State.Properties {
		if stored, found := properties[key]; found && !writable(&stored) {
			continue
		}
		properties[key] = schematypes.KnowledgeStateValue{Value: value, Unset: value == nil, KnowledgeEvidenceKey: k}
	}
	state.Properties = properties
	if evidence.StateMode == ke.StateModeReplace {
		state.Replaced = latestKey(state.Replaced, &k)
	}
	return state
}

// RebuildAliasState applies an alias's evidence in key order.
func RebuildAliasState(evidence []AliasEvidence) AliasRecord {
	ordered := slices.Clone(evidence)
	for i := range evidence {
		row := &evidence[i]
		if row.ReobservedAt != nil && row.ExpiresAt != nil {
			ordered = append(ordered, AliasEvidence{EffectiveAt: *row.ReobservedAt, ReobservationOf: row})
		}
	}
	slices.SortFunc(ordered, func(a, b AliasEvidence) int {
		return a.Key().Compare(b.Key())
	})
	var record AliasRecord
	for _, row := range ordered {
		record = ApplyEvidence(record, row)
	}
	return record
}

// keyBefore reports whether stored is absent or before k.
func keyBefore(stored *schematypes.KnowledgeEvidenceKey, k schematypes.KnowledgeEvidenceKey) bool {
	return stored == nil || stored.Compare(k) < 0
}

func latestKey(a, b *schematypes.KnowledgeEvidenceKey) *schematypes.KnowledgeEvidenceKey {
	if a == nil || (b != nil && b.Compare(*a) > 0) {
		return b
	}
	return a
}

func keyTime(k *schematypes.KnowledgeEvidenceKey) *time.Time {
	if k == nil {
		return nil
	}
	return &k.EffectiveAt
}

// SubjectAlias is one of a subject's aliases, as the subject merge reads it.
type SubjectAlias struct {
	ID        uuid.UUID
	Authority schematypes.KnowledgeAuthority
	Record    AliasRecord
}

// SubjectState is a subject's state merged from its aliases.
type SubjectState struct {
	State            schematypes.KnowledgeGraphSubjectState
	Sources          schematypes.KnowledgeStateSources
	StateEffectiveAt *time.Time
	LastObservedAt   *time.Time
	WithdrawnAt      *time.Time
	ExpiresAt        *time.Time
}

// MergeSubjectState merges a subject from its aliases. Each field comes from the highest authority tier with a
// live entry for it: the latest set value by evidence key, or absent when the tier only unsets it. A subject whose
// aliases are all withdrawn is withdrawn, and keeps the fields merged from all of them, so it still has a name.
func MergeSubjectState(aliases []SubjectAlias) SubjectState {
	var merged SubjectState
	live := make([]SubjectAlias, 0, len(aliases))
	for _, alias := range aliases {
		if alias.Record.WithdrawnAt == nil {
			live = append(live, alias)
		}
		merged.WithdrawnAt = latestTime(merged.WithdrawnAt, alias.Record.WithdrawnAt)
	}
	fieldAliases := live
	if len(live) > 0 {
		merged.WithdrawnAt = nil
	} else {
		fieldAliases = aliases
	}

	var displayName, description *fieldChoice
	properties := make(map[string]*fieldChoice)
	for _, alias := range fieldAliases {
		state := alias.Record.State
		displayName = displayName.consider(alias, state.DisplayName)
		description = description.consider(alias, state.Description)
		for key, value := range state.Properties {
			properties[key] = properties[key].consider(alias, &value)
		}
	}

	note := func(field *fieldChoice) {
		if !field.value.Unset {
			merged.StateEffectiveAt = latestTime(merged.StateEffectiveAt, &field.source.EffectiveAt)
		}
	}
	if displayName != nil {
		merged.State.DisplayName, _ = displayName.value.Value.(string)
		merged.Sources.DisplayName = &displayName.source
		note(displayName)
	}
	if description != nil {
		merged.State.Description, _ = description.value.Value.(string)
		merged.Sources.Description = &description.source
		note(description)
	}
	for key, property := range properties {
		if merged.Sources.Properties == nil {
			merged.Sources.Properties = make(map[string]schematypes.KnowledgeStateSource)
		}
		merged.Sources.Properties[key] = property.source
		if property.value.Unset {
			continue
		}
		if merged.State.Properties == nil {
			merged.State.Properties = make(map[string]any)
		}
		merged.State.Properties[key] = property.value.Value
		note(property)
	}

	neverExpires := false
	for _, alias := range live {
		merged.LastObservedAt = latestTime(merged.LastObservedAt, alias.Record.LastObservedAt)
		if alias.Record.ExpiresAt == nil {
			neverExpires = true
		}
		merged.ExpiresAt = latestTime(merged.ExpiresAt, alias.Record.ExpiresAt)
	}
	if neverExpires {
		merged.ExpiresAt = nil
	}
	return merged
}

// Equal reports whether two merged states store the same row. Times compare at the microsecond, the precision
// they are stored at, so a state read back from storage equals the state that was written.
func (s SubjectState) Equal(other SubjectState) bool {
	sameTime := func(a, b *time.Time) bool {
		if a == nil || b == nil {
			return a == b
		}
		return a.Truncate(time.Microsecond).Equal(b.Truncate(time.Microsecond))
	}
	sameJSON := func(a, b any) bool {
		encodedA, encodeErrA := json.Marshal(a)
		encodedB, encodeErrB := json.Marshal(b)
		return encodeErrA == nil && encodeErrB == nil && bytes.Equal(encodedA, encodedB)
	}
	return sameTime(s.StateEffectiveAt, other.StateEffectiveAt) &&
		sameTime(s.LastObservedAt, other.LastObservedAt) &&
		sameTime(s.WithdrawnAt, other.WithdrawnAt) &&
		sameTime(s.ExpiresAt, other.ExpiresAt) &&
		sameJSON(s.State, other.State) &&
		sameJSON(s.Sources, other.Sources)
}

// IsStale reports whether a subject or alias has expired by now.
func IsStale(expiresAt *time.Time, now time.Time) bool {
	return expiresAt != nil && expiresAt.Before(now)
}

func latestTime(a, b *time.Time) *time.Time {
	if a == nil || (b != nil && b.After(*a)) {
		return b
	}
	return a
}

// fieldChoice is the entry deciding one merged field so far, and the alias it came from.
type fieldChoice struct {
	value  schematypes.KnowledgeStateValue
	source schematypes.KnowledgeStateSource
}

// consider returns the better of the current choice and the alias's entry for the field. A higher tier always
// wins, set or unset. Within a tier a set value beats an unset, and otherwise the later evidence key wins, then
// the higher alias ID, which only decides between two aliases stating the field in one event.
func (current *fieldChoice) consider(alias SubjectAlias, value *schematypes.KnowledgeStateValue) *fieldChoice {
	if value == nil {
		return current
	}
	candidate := &fieldChoice{value: *value, source: schematypes.KnowledgeStateSource{
		AliasID: alias.ID, Authority: alias.Authority, EffectiveAt: value.EffectiveAt}}
	if current == nil {
		return candidate
	}
	isSet := func(choice *fieldChoice) int {
		if choice.value.Unset {
			return 0
		}
		return 1
	}
	better := cmp.Or(
		candidate.source.Authority.Compare(current.source.Authority),
		isSet(candidate)-isSet(current),
		candidate.value.Compare(current.value.KnowledgeEvidenceKey),
		bytes.Compare(candidate.source.AliasID[:], current.source.AliasID[:]),
	)
	if better > 0 {
		return candidate
	}
	return current
}
