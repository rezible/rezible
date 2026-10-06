package knowledgegraph

import (
	"bytes"
	"slices"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	kne "github.com/rezible/rezible/ent/knowledgeentity"
)

// StructureLink is one parent-to-child link of the structure hierarchy.
type StructureLink struct {
	Parent uuid.UUID
	Child  uuid.UUID
}

// StructureGraph is the part of the structure hierarchy a resolution needs.
type StructureGraph struct {
	Categories map[uuid.UUID]kne.Category
	Links      []StructureLink
}

// Resolve maps each entity to the entities in a target category that represent it. An entity in a target
// category represents itself. Any other entity climbs links, from child to parent, through intermediates of
// any category, for at most maxDepth steps; it is represented by every nearest ancestor in a target
// category. A cycle ends the climb. Entities represented by nothing are omitted, and each result is sorted
// by ID.
func Resolve(graph StructureGraph, entityIDs []uuid.UUID, targets []kne.Category, maxDepth int) map[uuid.UUID][]uuid.UUID {
	targetCategories := mapset.NewSet(targets...)
	parents := graph.parentsByChild()
	result := make(map[uuid.UUID][]uuid.UUID)
	for _, id := range entityIDs {
		var represented []uuid.UUID
		if graph.inCategory(id, targetCategories) {
			represented = []uuid.UUID{id}
		} else {
			represented = graph.climb(parents, id, targetCategories, maxDepth).ToSlice()
		}
		if len(represented) == 0 {
			continue
		}
		slices.SortFunc(represented, func(a uuid.UUID, b uuid.UUID) int {
			return bytes.Compare(a[:], b[:])
		})
		result[id] = represented
	}
	return result
}

func (g StructureGraph) IsLoaded(id uuid.UUID) bool {
	_, ok := g.Categories[id]
	return ok
}

func (g StructureGraph) inCategory(id uuid.UUID, categories mapset.Set[kne.Category]) bool {
	category, found := g.Categories[id]
	return found && categories.Contains(category)
}

func (g StructureGraph) parentsByChild() map[uuid.UUID]mapset.Set[uuid.UUID] {
	parents := make(map[uuid.UUID]mapset.Set[uuid.UUID])
	for _, link := range g.Links {
		childParents, found := parents[link.Child]
		if !found {
			childParents = mapset.NewSet[uuid.UUID]()
			parents[link.Child] = childParents
		}
		childParents.Add(link.Parent)
	}
	return parents
}

// climb walks up from the entity one step at a time and returns the ancestors in a target category found
// at the first step that reaches any. Each entity is visited once, so a cycle ends the climb.
func (g StructureGraph) climb(parents map[uuid.UUID]mapset.Set[uuid.UUID], id uuid.UUID, targets mapset.Set[kne.Category], maxDepth int) mapset.Set[uuid.UUID] {
	frontier := mapset.NewSet(id)
	visited := mapset.NewSet(id)
	reached := mapset.NewSet[uuid.UUID]()
	for depth := 0; depth < maxDepth && reached.IsEmpty() && !frontier.IsEmpty(); depth++ {
		next := mapset.NewSet[uuid.UUID]()
		for childID := range frontier.Iter() {
			childParents, found := parents[childID]
			if !found {
				continue
			}
			for parentID := range childParents.Iter() {
				if !visited.Add(parentID) {
					continue
				}
				next.Add(parentID)
				if g.inCategory(parentID, targets) {
					reached.Add(parentID)
				}
			}
		}
		frontier = next
	}
	return reached
}
