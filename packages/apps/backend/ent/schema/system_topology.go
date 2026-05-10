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

type SystemTopologySnapshot struct {
	ent.Schema
}

func (SystemTopologySnapshot) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SystemTopologySnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Time("as_of").Default(time.Now),
		field.String("name").Optional(),
		field.Enum("scope").
			Values("all", "incident").
			Default("all"),
		field.JSON("scope_properties", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemTopologySnapshot) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entities", SystemTopologySnapshotEntity.Type).
			Ref("snapshot"),
		edge.From("relationships", SystemTopologySnapshotRelationship.Type).
			Ref("snapshot"),
		edge.From("system_analyses", SystemAnalysis.Type).
			Ref("topology_snapshot"),
	}
}

func (SystemTopologySnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "as_of"),
		index.Fields("tenant_id", "created_at"),
	}
}

type SystemTopologySnapshotEntity struct {
	ent.Schema
}

func (SystemTopologySnapshotEntity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SystemTopologySnapshotEntity) Fields() []ent.Field {
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

func (SystemTopologySnapshotEntity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("snapshot", SystemTopologySnapshot.Type).
			Required().
			Unique().
			Field("snapshot_id"),
		edge.To("knowledge_entity", KnowledgeEntity.Type).
			Unique().
			Field("knowledge_entity_id"),
		edge.From("source_relationships", SystemTopologySnapshotRelationship.Type).
			Ref("source_snapshot_entity"),
		edge.From("target_relationships", SystemTopologySnapshotRelationship.Type).
			Ref("target_snapshot_entity"),
		edge.From("analysis_nodes", SystemAnalysisTopologyNode.Type).
			Ref("snapshot_entity"),
	}
}

func (SystemTopologySnapshotEntity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "snapshot_id"),
		index.Fields("tenant_id", "knowledge_entity_id"),
		index.Fields("tenant_id", "snapshot_id", "knowledge_entity_id").
			Unique(),
	}
}

type SystemTopologySnapshotRelationship struct {
	ent.Schema
}

func (SystemTopologySnapshotRelationship) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SystemTopologySnapshotRelationship) Fields() []ent.Field {
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

func (SystemTopologySnapshotRelationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("knowledge_relationship", KnowledgeRelationship.Type).
			Unique().
			Field("knowledge_relationship_id"),
		edge.To("snapshot", SystemTopologySnapshot.Type).
			Required().
			Unique().
			Field("snapshot_id"),
		edge.To("source_snapshot_entity", SystemTopologySnapshotEntity.Type).
			Required().
			Unique().
			Field("source_snapshot_entity_id"),
		edge.To("target_snapshot_entity", SystemTopologySnapshotEntity.Type).
			Required().
			Unique().
			Field("target_snapshot_entity_id"),
		edge.From("analysis_edges", SystemAnalysisTopologyEdge.Type).
			Ref("snapshot_relationship"),
	}
}

func (SystemTopologySnapshotRelationship) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "snapshot_id"),
		index.Fields("tenant_id", "knowledge_relationship_id"),
		index.Fields("tenant_id", "source_snapshot_entity_id"),
		index.Fields("tenant_id", "target_snapshot_entity_id"),
		index.Fields("tenant_id", "snapshot_id", "knowledge_relationship_id").
			Unique(),
	}
}
