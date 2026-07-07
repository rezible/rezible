package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type KnowledgeGraphSnapshot struct {
	ent.Schema
}

func (KnowledgeGraphSnapshot) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (KnowledgeGraphSnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Time("as_of").Default(time.Now),
		field.String("name").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (KnowledgeGraphSnapshot) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("system_analyses", SystemAnalysis.Type).
			Ref("knowledge_graph_snapshot"),
		edge.From("entities", KnowledgeGraphSnapshotEntity.Type).
			Ref("snapshot"),
		edge.From("relationships", KnowledgeGraphSnapshotRelationship.Type).
			Ref("snapshot"),
	}
}

func (KnowledgeGraphSnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "as_of"),
		index.Fields("tenant_id", "created_at"),
	}
}

type KnowledgeGraphSnapshotEntity struct {
	ent.Schema
}

func (KnowledgeGraphSnapshotEntity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (KnowledgeGraphSnapshotEntity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("snapshot_id", uuid.UUID{}),
		field.UUID("knowledge_entity_id", uuid.UUID{}).Optional().Nillable(),
		field.String("entity_kind").NotEmpty(),
		field.String("display_name").NotEmpty(),
		field.Text("description").Optional(),
		field.JSON("properties", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("aliases", []map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("created_at").Default(time.Now),
	}
}

func (KnowledgeGraphSnapshotEntity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("snapshot", KnowledgeGraphSnapshot.Type).
			Required().
			Unique().
			Field("snapshot_id"),
		edge.To("knowledge_entity", KnowledgeEntity.Type).
			Unique().
			Field("knowledge_entity_id"),
		edge.From("source_relationships", KnowledgeGraphSnapshotRelationship.Type).
			Ref("source_snapshot_entity"),
		edge.From("target_relationships", KnowledgeGraphSnapshotRelationship.Type).
			Ref("target_snapshot_entity"),
	}
}

func (KnowledgeGraphSnapshotEntity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "snapshot_id"),
		index.Fields("tenant_id", "knowledge_entity_id"),
		index.Fields("tenant_id", "snapshot_id", "knowledge_entity_id").
			Unique(),
	}
}

type KnowledgeGraphSnapshotRelationship struct {
	ent.Schema
}

func (KnowledgeGraphSnapshotRelationship) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (KnowledgeGraphSnapshotRelationship) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("snapshot_id", uuid.UUID{}),
		field.UUID("knowledge_relationship_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("source_snapshot_entity_id", uuid.UUID{}),
		field.UUID("target_snapshot_entity_id", uuid.UUID{}),
		field.String("relationship_kind").NotEmpty(),
		field.String("display_name").Optional(),
		field.Text("description").Optional(),
		field.JSON("properties", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("created_at").Default(time.Now),
	}
}

func (KnowledgeGraphSnapshotRelationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("knowledge_relationship", KnowledgeRelationship.Type).
			Unique().
			Field("knowledge_relationship_id"),
		edge.To("snapshot", KnowledgeGraphSnapshot.Type).
			Required().
			Unique().
			Field("snapshot_id"),
		edge.To("source_snapshot_entity", KnowledgeGraphSnapshotEntity.Type).
			Required().
			Unique().
			Field("source_snapshot_entity_id"),
		edge.To("target_snapshot_entity", KnowledgeGraphSnapshotEntity.Type).
			Required().
			Unique().
			Field("target_snapshot_entity_id"),
		edge.From("analysis_edges", SystemAnalysisTopologyEdge.Type).
			Ref("snapshot_relationship"),
	}
}

func (KnowledgeGraphSnapshotRelationship) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "snapshot_id"),
		index.Fields("tenant_id", "knowledge_relationship_id"),
		index.Fields("tenant_id", "source_snapshot_entity_id"),
		index.Fields("tenant_id", "target_snapshot_entity_id"),
		index.Fields("tenant_id", "snapshot_id", "knowledge_relationship_id").
			Unique(),
	}
}
