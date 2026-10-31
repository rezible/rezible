package knowledgegraph

import (
	kne "github.com/rezible/rezible/ent/knowledgeentity"
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
