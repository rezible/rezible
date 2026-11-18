package schematypes

import (
	"bytes"
	"cmp"
	"database/sql/driver"
	"fmt"
	"maps"
	"slices"
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

// KnowledgeEvidenceKey orders evidence on one alias: by effective time, then event ID. Evidence is unique per
// event and alias, so the order is total.
type KnowledgeEvidenceKey struct {
	EffectiveAt time.Time `json:"effective_at"`
	EventID     uuid.UUID `json:"event_id"`
}

// Compare orders keys by effective time, then by the bytes of the event ID.
func (k KnowledgeEvidenceKey) Compare(other KnowledgeEvidenceKey) int {
	return cmp.Or(k.EffectiveAt.Compare(other.EffectiveAt), bytes.Compare(k.EventID[:], other.EventID[:]))
}

// KnowledgeStateValue is one field of an alias's state: a value, or a tombstone when Unset, with the key of
// the evidence that stated it.
type KnowledgeStateValue struct {
	Value any  `json:"value,omitempty"`
	Unset bool `json:"unset,omitempty"`
	KnowledgeEvidenceKey
}

// KnowledgeAliasState is what one alias's evidence says, field by field.
type KnowledgeAliasState struct {
	DisplayName *KnowledgeStateValue           `json:"display_name,omitempty"`
	Description *KnowledgeStateValue           `json:"description,omitempty"`
	Properties  map[string]KnowledgeStateValue `json:"properties,omitempty"`
	// Replaced is the key of the latest replace evidence; property statements before it are ignored.
	Replaced *KnowledgeEvidenceKey `json:"replaced,omitempty"`
	// LastObserved is the key of the latest observation.
	LastObserved *KnowledgeEvidenceKey `json:"last_observed,omitempty"`
	// Withdrawn is the key of the latest deletion when no observation follows it.
	Withdrawn *KnowledgeEvidenceKey `json:"withdrawn,omitempty"`
}

// KnowledgeAuthority is how authoritative an alias's source is for its subject, ordered from lowest to
// highest. It is a string everywhere but in the database, where Value and Scan store it as its rank.
type KnowledgeAuthority string

const (
	// KnowledgeAuthorityObserved is a source that mentions the subject while reporting something else.
	KnowledgeAuthorityObserved KnowledgeAuthority = "observed"
	// KnowledgeAuthorityDeclared is a source that is a system of record for the subject.
	KnowledgeAuthorityDeclared KnowledgeAuthority = "declared"
)

// knowledgeAuthorityRanks are the values the authority column stores. Their numeric order is the authority
// order, which SQL ordering and comparisons on the column rely on. The gaps leave room for a new tier between or
// above the others without renumbering.
var knowledgeAuthorityRanks = map[KnowledgeAuthority]int16{
	KnowledgeAuthorityObserved: 10,
	KnowledgeAuthorityDeclared: 20,
}

// IsValid reports whether the authority is one of the defined tiers.
func (a KnowledgeAuthority) IsValid() bool {
	_, defined := knowledgeAuthorityRanks[a]
	return defined
}

// Compare orders authorities from lowest to highest, like cmp.Compare. An undefined authority is lowest.
func (a KnowledgeAuthority) Compare(other KnowledgeAuthority) int {
	return cmp.Compare(knowledgeAuthorityRanks[a], knowledgeAuthorityRanks[other])
}

// Rank returns the value the authority column stores for the authority.
func (a KnowledgeAuthority) Rank() (int16, error) {
	rank, defined := knowledgeAuthorityRanks[a]
	if !defined {
		return 0, fmt.Errorf("undefined knowledge authority %q", a)
	}
	return rank, nil
}

// KnowledgeAuthorityOfRank returns the authority the authority column's value stores.
func KnowledgeAuthorityOfRank(rank int64) (KnowledgeAuthority, error) {
	for authority, authorityRank := range knowledgeAuthorityRanks {
		if int64(authorityRank) == rank {
			return authority, nil
		}
	}
	return "", fmt.Errorf("undefined knowledge authority rank %d", rank)
}

// Value stores the authority as its rank, so SQL orders and compares authorities in authority order.
func (a KnowledgeAuthority) Value() (driver.Value, error) {
	rank, rankErr := a.Rank()
	if rankErr != nil {
		return nil, rankErr
	}
	return int64(rank), nil
}

// Scan reads an authority stored as its rank.
func (a *KnowledgeAuthority) Scan(src any) error {
	rank, isRank := src.(int64)
	if !isRank {
		return fmt.Errorf("scan knowledge authority: %T is not a rank", src)
	}
	authority, authorityErr := KnowledgeAuthorityOfRank(rank)
	if authorityErr != nil {
		return authorityErr
	}
	*a = authority
	return nil
}

// KnowledgeAuthorityRanks returns every stored authority value, in authority order.
func KnowledgeAuthorityRanks() []int16 {
	ranks := slices.Collect(maps.Values(knowledgeAuthorityRanks))
	slices.Sort(ranks)
	return ranks
}

// KnowledgeStateSource is the alias that supplied a subject's field, with the alias's authority and the
// field's effective time. For a field absent because its deciding tier only unsets it, it is the unsetting alias.
type KnowledgeStateSource struct {
	AliasID     uuid.UUID          `json:"alias_id"`
	Authority   KnowledgeAuthority `json:"authority"`
	EffectiveAt time.Time          `json:"effective_at"`
}

// KnowledgeStateSources records, for each field of a subject's merged state, where it came from.
type KnowledgeStateSources struct {
	DisplayName *KnowledgeStateSource           `json:"display_name,omitempty"`
	Description *KnowledgeStateSource           `json:"description,omitempty"`
	Properties  map[string]KnowledgeStateSource `json:"properties,omitempty"`
}

type SkippedEvidenceReason string

const (
	SkippedEvidenceReasonPredicateCategories SkippedEvidenceReason = "predicate_categories"
	SkippedEvidenceReasonSelfContains        SkippedEvidenceReason = "self_contains"
	SkippedEvidenceReasonTopologyMismatch    SkippedEvidenceReason = "topology_mismatch"
	SkippedEvidenceReasonMissingEndpoint     SkippedEvidenceReason = "missing_endpoint"
)

// SkippedEvidenceRef records relationship evidence omitted from a provider projection.
type SkippedEvidenceRef struct {
	Reason      SkippedEvidenceReason `json:"reason"`
	ResourceRef string                `json:"resource_ref"`
	Error       string                `json:"error"`
}
