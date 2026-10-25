package extensions

import (
	"testing"

	"entgo.io/ent"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/rezible/rezible/ent/schema/annotations"
)

// KnowledgeEntity stands in for the knowledge entity schema; its table is knowledge_entities. Tagged makes
// Tagged's edge to it many-to-many.
type KnowledgeEntity struct {
	ent.Schema
	tagged bool
}

func (KnowledgeEntity) Fields() []ent.Field {
	return []ent.Field{field.UUID("id", uuid.UUID{})}
}

func (e KnowledgeEntity) Edges() []ent.Edge {
	if !e.tagged {
		return nil
	}
	return []ent.Edge{edge.From("tagged", Tagged.Type).Ref("entities")}
}

// KnowledgeRelationship stands in for the knowledge relationship schema; its table is knowledge_relationships.
type KnowledgeRelationship struct{ ent.Schema }

func (KnowledgeRelationship) Fields() []ent.Field {
	return []ent.Field{field.UUID("id", uuid.UUID{})}
}

// Link refers to a knowledge relationship, so the relationship list is not empty.
type Link struct{ ent.Schema }

func (Link) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}),
		field.UUID("knowledge_relationship_id", uuid.UUID{}),
	}
}

func (Link) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("knowledge_relationship", KnowledgeRelationship.Type).Field("knowledge_relationship_id").Required().Unique(),
	}
}

// Analysis scopes the subjects below.
type Analysis struct{ ent.Schema }

func (Analysis) Fields() []ent.Field {
	return []ent.Field{field.UUID("id", uuid.UUID{})}
}

// Subject refers to a knowledge entity through an edge field of its own, re-pointed or deduplicated as the
// test configures it.
type Subject struct {
	ent.Schema
	immutable bool
	annotated bool
	unique    bool
}

func (s Subject) Fields() []ent.Field {
	entityID := field.UUID("knowledge_entity_id", uuid.UUID{})
	if s.immutable {
		entityID = entityID.Immutable()
	}
	return []ent.Field{
		field.UUID("id", uuid.UUID{}),
		field.UUID("analysis_id", uuid.UUID{}),
		entityID,
	}
}

func (s Subject) Edges() []ent.Edge {
	entity := edge.To("knowledge_entity", KnowledgeEntity.Type).Field("knowledge_entity_id").Required().Unique()
	if s.immutable {
		entity = entity.Immutable()
	}
	if s.annotated {
		entity = entity.Annotations(annotations.EntityMergeDeletesDuplicateRows())
	}
	return []ent.Edge{
		edge.To("analysis", Analysis.Type).Field("analysis_id").Required().Unique(),
		entity,
	}
}

func (s Subject) Indexes() []ent.Index {
	if !s.unique {
		return nil
	}
	return []ent.Index{index.Fields("analysis_id", "knowledge_entity_id").Unique()}
}

// Tagged refers to knowledge entities through a join table.
type Tagged struct{ ent.Schema }

func (Tagged) Fields() []ent.Field {
	return []ent.Field{field.UUID("id", uuid.UUID{})}
}

func (Tagged) Edges() []ent.Edge {
	return []ent.Edge{edge.To("entities", KnowledgeEntity.Type)}
}

func collectSchemas(t *testing.T, schemas ...ent.Interface) (*references, error) {
	t.Helper()
	return collectWithCore(t, nil, schemas...)
}

func collectWithCore(t *testing.T, core map[column]bool, schemas ...ent.Interface) (*references, error) {
	t.Helper()
	loaded := make([]*load.Schema, len(schemas))
	for i, s := range schemas {
		encoded, marshalErr := load.MarshalSchema(s)
		require.NoError(t, marshalErr)
		decoded, unmarshalErr := load.UnmarshalSchema(encoded)
		require.NoError(t, unmarshalErr)
		loaded[i] = decoded
	}
	storage, storageErr := gen.NewStorage("sql")
	require.NoError(t, storageErr)
	graph, graphErr := gen.NewGraph(&gen.Config{Package: "example/ent", Storage: storage}, loaded...)
	require.NoError(t, graphErr)
	return collect(graph, core)
}

func TestCollectListsReferences(t *testing.T) {
	refs, collectErr := collectSchemas(t, KnowledgeEntity{}, KnowledgeRelationship{}, Link{}, Analysis{}, Subject{annotated: true, unique: true})

	require.NoError(t, collectErr)
	require.Len(t, refs.Entities, 1)
	require.Equal(t, reference{
		Table:            "subjects",
		Column:           "knowledge_entity_id",
		Node:             "Subject",
		Setter:           "SetKnowledgeEntityID",
		DeleteDuplicates: true,
		UniqueWith:       []string{"analysis_id"},
	}, refs.Entities[0])
	require.Equal(t, []string{
		"links_knowledge_relationships_knowledge_relationship",
		"subjects_knowledge_entities_knowledge_entity",
	}, refs.ForeignKeyConstraints)

	plain, plainErr := collectSchemas(t, KnowledgeEntity{}, KnowledgeRelationship{}, Link{}, Analysis{}, Subject{unique: true})
	require.NoError(t, plainErr)
	require.Len(t, plain.Entities, 1)
	require.False(t, plain.Entities[0].DeleteDuplicates, "a unique column re-points by default")
}

func TestCollectRejectsUnsupportedReferences(t *testing.T) {
	cases := map[string]struct {
		schemas  []ent.Interface
		contains string
	}{
		"deleting duplicates of a column outside a unique key": {
			schemas:  []ent.Interface{KnowledgeEntity{}, KnowledgeRelationship{}, Link{}, Analysis{}, Subject{annotated: true}},
			contains: "is not part of a unique key",
		},
		"an immutable reference": {
			schemas:  []ent.Interface{KnowledgeEntity{}, KnowledgeRelationship{}, Link{}, Analysis{}, Subject{immutable: true}},
			contains: "is Immutable()",
		},
		"a missing subject type": {
			schemas:  []ent.Interface{KnowledgeEntity{}, Analysis{}, Subject{}},
			contains: "schema type KnowledgeRelationship is missing",
		},
		"a subject without references": {
			schemas:  []ent.Interface{KnowledgeEntity{}, KnowledgeRelationship{}, Analysis{}, Subject{}},
			contains: "found 1 entity and 0 relationship references",
		},
		"a join table": {
			schemas:  []ent.Interface{KnowledgeEntity{tagged: true}, KnowledgeRelationship{}, Link{}, Tagged{}},
			contains: "neither a core column nor an edge of its own table",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, collectErr := collectSchemas(t, tc.schemas...)
			require.ErrorContains(t, collectErr, tc.contains)
		})
	}
}

func TestCollectRejectsAStaleCoreColumn(t *testing.T) {
	schemas := []ent.Interface{KnowledgeEntity{}, KnowledgeRelationship{}, Link{}, Analysis{}, Subject{}}
	found := map[column]bool{{"links", "knowledge_relationship_id"}: true}
	stale := map[column]bool{{"links", "renamed_relationship_id"}: true}

	_, foundErr := collectWithCore(t, found, schemas...)
	_, staleErr := collectWithCore(t, stale, schemas...)

	require.ErrorContains(t, foundErr, "0 relationship references", "a core column is not generated")
	require.ErrorContains(t, staleErr, "core column links.renamed_relationship_id is no longer a foreign key")
}
