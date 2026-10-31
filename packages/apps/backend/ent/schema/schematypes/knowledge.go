package schematypes

import (
	"time"

	"github.com/google/uuid"
)

type KnowledgeGraphSubjectState struct {
	DisplayName string         `json:"display_name,omitempty"`
	Description string         `json:"description,omitempty"`
	Properties  map[string]any `json:"properties"`
}

// IsEmpty reports whether the state has no display name, description or properties.
func (s KnowledgeGraphSubjectState) IsEmpty() bool {
	return s.DisplayName == "" && s.Description == "" && len(s.Properties) == 0
}

// KnowledgeStateValue is one field of an alias's state: a value, or a tombstone when Unset, with the effective
// time of the evidence that stated it.
type KnowledgeStateValue struct {
	Value       any       `json:"value,omitempty"`
	Unset       bool      `json:"unset,omitempty"`
	EffectiveAt time.Time `json:"effective_at"`
}

// KnowledgeAliasState is what one alias's evidence says, field by field.
type KnowledgeAliasState struct {
	DisplayName *KnowledgeStateValue           `json:"display_name,omitempty"`
	Description *KnowledgeStateValue           `json:"description,omitempty"`
	Properties  map[string]KnowledgeStateValue `json:"properties,omitempty"`
	// ReplacedAt is the effective time of the latest replace evidence; older statements are ignored.
	ReplacedAt *time.Time `json:"replaced_at,omitempty"`
}

// KnowledgeStateSource is the alias that supplied a subject's field, and the field's effective time.
type KnowledgeStateSource struct {
	AliasID     uuid.UUID `json:"alias_id"`
	EffectiveAt time.Time `json:"effective_at"`
}

// KnowledgeStateSources records, for each field of a subject's merged state, where it came from.
type KnowledgeStateSources struct {
	DisplayName *KnowledgeStateSource           `json:"display_name,omitempty"`
	Description *KnowledgeStateSource           `json:"description,omitempty"`
	Properties  map[string]KnowledgeStateSource `json:"properties,omitempty"`
}
