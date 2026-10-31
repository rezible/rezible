package knowledgegraph

import (
	"bytes"
	"cmp"
	"slices"
	"time"

	"github.com/google/uuid"

	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	"github.com/rezible/rezible/ent/schema/schematypes"
)

// AliasEvidence is one evidence row as alias state reads it.
type AliasEvidence struct {
	Kind        ke.Kind
	StateMode   ke.StateMode
	EffectiveAt time.Time
	ExpiresAt   *time.Time
	State       schematypes.KnowledgeGraphSubjectState
}

// AliasRecord is an alias's state and observation times.
type AliasRecord struct {
	State           schematypes.KnowledgeAliasState
	FirstObservedAt *time.Time
	LastObservedAt  *time.Time
	WithdrawnAt     *time.Time
	ExpiresAt       *time.Time
}

// ApplyEvidence returns the alias record after one more evidence row. Applying rows in any order gives the
// result of RebuildAliasState over them, except that rows with equal effective times are taken in the order
// applied.
func ApplyEvidence(record AliasRecord, evidence AliasEvidence) AliasRecord {
	t := evidence.EffectiveAt
	if evidence.Kind == ke.KindDeleted {
		if record.LastObservedAt == nil || !record.LastObservedAt.After(t) {
			record.WithdrawnAt = latestTime(record.WithdrawnAt, &t)
		}
		return record
	}

	record.State = applyStatements(record.State, evidence)
	if record.LastObservedAt == nil || !t.Before(*record.LastObservedAt) {
		record.ExpiresAt = evidence.ExpiresAt
	}
	if record.FirstObservedAt == nil || t.Before(*record.FirstObservedAt) {
		record.FirstObservedAt = &t
	}
	record.LastObservedAt = latestTime(record.LastObservedAt, &t)
	if record.WithdrawnAt != nil && !record.WithdrawnAt.After(t) {
		record.WithdrawnAt = nil
	}
	return record
}

func applyStatements(state schematypes.KnowledgeAliasState, evidence AliasEvidence) schematypes.KnowledgeAliasState {
	t := evidence.EffectiveAt
	writable := func(stored *schematypes.KnowledgeStateValue) bool {
		return stored == nil || !stored.EffectiveAt.After(t)
	}
	if name := evidence.State.DisplayName; name != "" && writable(state.DisplayName) {
		state.DisplayName = &schematypes.KnowledgeStateValue{Value: name, EffectiveAt: t}
	}
	if description := evidence.State.Description; description != "" && writable(state.Description) {
		state.Description = &schematypes.KnowledgeStateValue{Value: description, EffectiveAt: t}
	}
	// Only properties are ignored before the latest replace: a replace unsets the properties it omits but never
	// the display name or description, so ignoring those too would depend on arrival order.
	if state.ReplacedAt != nil && t.Before(*state.ReplacedAt) {
		return state
	}
	properties := make(map[string]schematypes.KnowledgeStateValue, len(state.Properties)+len(evidence.State.Properties))
	for key, stored := range state.Properties {
		_, stated := evidence.State.Properties[key]
		if evidence.StateMode == ke.StateModeReplace && !stated && !stored.EffectiveAt.After(t) {
			stored = schematypes.KnowledgeStateValue{Unset: true, EffectiveAt: t}
		}
		properties[key] = stored
	}
	for key, value := range evidence.State.Properties {
		if stored, found := properties[key]; found && !writable(&stored) {
			continue
		}
		properties[key] = schematypes.KnowledgeStateValue{Value: value, Unset: value == nil, EffectiveAt: t}
	}
	state.Properties = properties
	if evidence.StateMode == ke.StateModeReplace {
		state.ReplacedAt = latestTime(state.ReplacedAt, &t)
	}
	return state
}

// RebuildAliasState applies an alias's evidence, given in ingestion order (created at, then ID), ordered by
// effective time.
func RebuildAliasState(evidence []AliasEvidence) AliasRecord {
	ordered := slices.Clone(evidence)
	slices.SortStableFunc(ordered, func(a, b AliasEvidence) int {
		return a.EffectiveAt.Compare(b.EffectiveAt)
	})
	var record AliasRecord
	for _, row := range ordered {
		record = ApplyEvidence(record, row)
	}
	return record
}

// SubjectAlias is one of a subject's aliases, as the subject merge reads it.
type SubjectAlias struct {
	ID        uuid.UUID
	CreatedAt time.Time
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

// MergeSubjectState merges a subject from its aliases. Each field takes the latest set value among the live
// aliases; ties go to the alias created later, then the higher ID. A subject is withdrawn when it has aliases
// and all are withdrawn; it then keeps the fields merged from all of them, so a withdrawn subject still reads
// with its last known name.
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

	// Later aliases win ties, so visit them last.
	ordered := slices.Clone(fieldAliases)
	slices.SortFunc(ordered, func(a, b SubjectAlias) int {
		return cmp.Or(a.CreatedAt.Compare(b.CreatedAt), bytes.Compare(a.ID[:], b.ID[:]))
	})
	type chosen struct {
		value  schematypes.KnowledgeStateValue
		source schematypes.KnowledgeStateSource
	}
	choose := func(current *chosen, aliasID uuid.UUID, value *schematypes.KnowledgeStateValue) *chosen {
		if value == nil || value.Unset || (current != nil && value.EffectiveAt.Before(current.value.EffectiveAt)) {
			return current
		}
		return &chosen{value: *value, source: schematypes.KnowledgeStateSource{AliasID: aliasID, EffectiveAt: value.EffectiveAt}}
	}
	var displayName, description *chosen
	properties := make(map[string]*chosen)
	for _, alias := range ordered {
		state := alias.Record.State
		displayName = choose(displayName, alias.ID, state.DisplayName)
		description = choose(description, alias.ID, state.Description)
		for key, value := range state.Properties {
			properties[key] = choose(properties[key], alias.ID, &value)
		}
	}

	note := func(field *chosen) {
		merged.StateEffectiveAt = latestTime(merged.StateEffectiveAt, &field.source.EffectiveAt)
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
		if property == nil {
			continue
		}
		if merged.State.Properties == nil {
			merged.State.Properties = make(map[string]any)
			merged.Sources.Properties = make(map[string]schematypes.KnowledgeStateSource)
		}
		merged.State.Properties[key] = property.value.Value
		merged.Sources.Properties[key] = property.source
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
