package knowledgegraph

import (
	"bytes"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	kne "github.com/rezible/rezible/ent/knowledgeentity"
)

// testGraph builds a structure graph from entity names, so cases can name entities instead of IDs.
type testGraph struct {
	ids   map[string]uuid.UUID
	graph StructureGraph
}

func newTestGraph(categories map[string]kne.Category, links [][2]string) *testGraph {
	g := &testGraph{
		ids: make(map[string]uuid.UUID),
		graph: StructureGraph{
			Categories: make(map[uuid.UUID]kne.Category),
		},
	}
	for name, category := range categories {
		g.graph.Categories[g.id(name)] = category
	}
	for _, pair := range links {
		link := StructureLink{
			Parent: g.id(pair[0]),
			Child:  g.id(pair[1]),
		}
		g.graph.Links = append(g.graph.Links, link)
	}
	return g
}

func (g *testGraph) id(name string) uuid.UUID {
	id, found := g.ids[name]
	if !found {
		id = uuid.New()
		g.ids[name] = id
	}
	return id
}

func (g *testGraph) sortedIDs(names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(names))
	for _, name := range names {
		ids = append(ids, g.id(name))
	}
	slices.SortFunc(ids, func(a uuid.UUID, b uuid.UUID) int {
		return bytes.Compare(a[:], b[:])
	})
	return ids
}

func TestResolve(t *testing.T) {
	g := newTestGraph(
		map[string]kne.Category{
			"checkout":   kne.CategoryContainer,
			"payments":   kne.CategoryContainer,
			"billing":    kne.CategoryContainer,
			"search":     kne.CategoryContainer,
			"cluster":    kne.CategoryInfrastructure,
			"shop":       kne.CategorySystem,
			"runbook":    kne.CategoryProcess,
			"cart":       kne.CategoryComponent,
			"ledger":     kne.CategoryComponent,
			"receipts":   kne.CategoryComponent,
			"pdf":        kne.CategoryCode,
			"orders":     kne.CategoryComponent,
			"pricing":    kne.CategoryComponent,
			"discounts":  kne.CategoryCode,
			"telemetry":  kne.CategoryCode,
			"loop-start": kne.CategoryComponent,
			"loop-end":   kne.CategoryComponent,
		},
		[][2]string{
			{"shop", "checkout"},
			{"checkout", "cart"},
			{"checkout", "ledger"},
			{"payments", "ledger"},
			{"billing", "receipts"},
			{"receipts", "pdf"},
			{"payments", "pdf"},
			{"checkout", "orders"},
			{"orders", "pricing"},
			{"pricing", "discounts"},
			{"checkout", "telemetry"},
			{"payments", "telemetry"},
			{"search", "telemetry"},
			{"cluster", "telemetry"},
			{"checkout", "runbook"},
			{"loop-start", "loop-end"},
			{"loop-end", "loop-start"},
		},
	)

	tests := []struct {
		name     string
		entity   string
		level    StructureLevel
		maxDepth int
		expected []string
	}{
		{
			name:     "a container in a target category is itself",
			entity:   "checkout",
			level:    StructureLevelRuntime,
			maxDepth: 3,
			expected: []string{"checkout"},
		},
		{
			name:     "a component resolves to its container",
			entity:   "cart",
			level:    StructureLevelRuntime,
			maxDepth: 3,
			expected: []string{"checkout"},
		},
		{
			name:     "a component resolves to several nearest containers",
			entity:   "ledger",
			level:    StructureLevelRuntime,
			maxDepth: 3,
			expected: []string{"checkout", "payments"},
		},
		{
			name:     "a farther container is not nearest",
			entity:   "pdf",
			level:    StructureLevelRuntime,
			maxDepth: 3,
			expected: []string{"payments"},
		},
		{
			name:     "the climb passes intermediates of other categories",
			entity:   "discounts",
			level:    StructureLevelRuntime,
			maxDepth: 3,
			expected: []string{"checkout"},
		},
		{
			name:     "the climb stops at max depth",
			entity:   "discounts",
			level:    StructureLevelRuntime,
			maxDepth: 2,
			expected: nil,
		},
		{
			name:     "a cycle ends the climb",
			entity:   "loop-start",
			level:    StructureLevelRuntime,
			maxDepth: 10,
			expected: nil,
		},
		{
			name:     "an entity above the level resolves to nothing",
			entity:   "shop",
			level:    StructureLevelRuntime,
			maxDepth: 3,
			expected: nil,
		},
		{
			name:     "an unmapped entity climbs like any other",
			entity:   "runbook",
			level:    StructureLevelRuntime,
			maxDepth: 3,
			expected: []string{"checkout"},
		},
		{
			name:     "an entity missing from the graph resolves to nothing",
			entity:   "unknown",
			level:    StructureLevelRuntime,
			maxDepth: 3,
			expected: nil,
		},
		{
			name:     "results are sorted by ID",
			entity:   "telemetry",
			level:    StructureLevelRuntime,
			maxDepth: 3,
			expected: []string{"checkout", "payments", "search", "cluster"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entityID := g.id(tc.entity)

			resolved := Resolve(g.graph, []uuid.UUID{entityID}, tc.level.Categories(), tc.maxDepth)

			if tc.expected == nil {
				assert.Empty(t, resolved)
				return
			}
			assert.Equal(t, map[uuid.UUID][]uuid.UUID{entityID: g.sortedIDs(tc.expected...)}, resolved)
		})
	}
}
