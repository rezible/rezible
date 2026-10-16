package knowledgegraph

import (
	"slices"

	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
)

// Knowledge entity kinds that providers link by name: a service is a container, a repository is code.
const (
	EntityKindService    = "service"
	EntityKindRepository = "repository"
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

// StructureLevel is a level of a system's structure, from the broadest (landscape) to the most detailed (implementation).
type StructureLevel int

const (
	StructureLevelLandscape StructureLevel = iota
	StructureLevelSystems
	StructureLevelRuntime
	StructureLevelImplementation
)

var structureLevelCategories = map[StructureLevel][]kne.Category{
	StructureLevelLandscape:      {kne.CategorySystemFunction},
	StructureLevelSystems:        {kne.CategorySystem},
	StructureLevelRuntime:        {kne.CategoryContainer, kne.CategoryInfrastructure},
	StructureLevelImplementation: {kne.CategoryComponent, kne.CategoryCode},
}

func (l StructureLevel) Categories() []kne.Category {
	return slices.Clone(structureLevelCategories[l])
}

// StructuralPredicates are the relationship predicates that express the structure hierarchy, from parent (source) to child (target).
var StructuralPredicates = []knr.Predicate{
	knr.PredicateContains,
}

// RelationshipClass groups relationship predicates by how they relate their endpoints.
type RelationshipClass int

const (
	// RelationshipClassDependsOn relationships mean the source depends on the target.
	RelationshipClassDependsOn RelationshipClass = iota
	// RelationshipClassAdjacent relationships relate their endpoints in either direction.
	RelationshipClassAdjacent
)

var relationshipClassPredicates = map[RelationshipClass][]knr.Predicate{
	RelationshipClassDependsOn: {
		knr.PredicateCalls,
		knr.PredicateReadsFrom,
		knr.PredicateWritesTo,
		knr.PredicateDependsOn,
		knr.PredicateRunsOn,
		knr.PredicateUses,
	},
	RelationshipClassAdjacent: {
		knr.PredicateConsumesFrom,
		knr.PredicatePublishesTo,
		knr.PredicateInteractsWith,
	},
}

// Predicates is the class's predicate set.
func (c RelationshipClass) Predicates() []knr.Predicate {
	return slices.Clone(relationshipClassPredicates[c])
}
