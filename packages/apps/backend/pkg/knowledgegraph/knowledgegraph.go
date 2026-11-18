package knowledgegraph

import (
	"bytes"
	"cmp"
	"time"

	"github.com/google/uuid"

	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/schema/schematypes"
)

// Knowledge entity kinds Rezible reasons about. Every other kind is a free-form refinement of its category.
const (
	EntityKindService      = "service"
	EntityKindRepository   = "repository"
	EntityKindUser         = "user"
	EntityKindTeam         = "team"
	EntityKindAlert        = "alert"
	EntityKindSystemHazard = "system_hazard"
)

// IsService reports whether the category and kind name a service.
func IsService(category kne.Category, kind string) bool {
	return category == kne.CategoryContainer && kind == EntityKindService
}

// IsMergeable reports whether a confirmed match can merge entities of the category and kind: services and
// repositories.
func IsMergeable(category kne.Category, kind string) bool {
	return IsService(category, kind) || (category == kne.CategoryCode && kind == EntityKindRepository)
}

// ParentClaim is a current contains relationship targeting an entity, as RankParentClaims ranks it: the highest
// authority and the earliest first observation among the relationship's live aliases.
type ParentClaim struct {
	RelationshipID  uuid.UUID
	Authority       schematypes.KnowledgeAuthority
	FirstObservedAt time.Time
}

// RankParentClaims orders an entity's parent claims best first: the highest authority, then the earliest first
// observation, then the lowest ID. All three come from stored evidence, so the order does not depend on the
// order claims were processed in. Placement chooses the best claim that closes no cycle. Staleness does not
// affect the rank.
func RankParentClaims(a, b ParentClaim) int {
	return cmp.Or(
		b.Authority.Compare(a.Authority),
		a.FirstObservedAt.Compare(b.FirstObservedAt),
		bytes.Compare(a.RelationshipID[:], b.RelationshipID[:]),
	)
}
