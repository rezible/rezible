package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type SystemAnalysis struct {
	ent.Schema
}

func (SystemAnalysis) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SystemAnalysis) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
	}
}

func (SystemAnalysis) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("analysis_nodes", SystemAnalysisTopologyNode.Type).
			Ref("analysis"),
		edge.From("analysis_edges", SystemAnalysisTopologyEdge.Type).
			Ref("analysis"),
	}
}

type SystemAnalysisTopologyNode struct {
	ent.Schema
}

func (SystemAnalysisTopologyNode) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SystemAnalysisTopologyNode) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("analysis_id", uuid.UUID{}),
		field.UUID("knowledge_entity_id", uuid.UUID{}),
		field.Time("referenced_at").Default(time.Now),
		field.Text("description").Optional(),
		field.Float("pos_x").Default(0),
		field.Float("pos_y").Default(0),
	}
}

func (SystemAnalysisTopologyNode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("analysis", SystemAnalysis.Type).
			Required().
			Unique().
			Field("analysis_id"),
		edge.To("knowledge_entity", KnowledgeEntity.Type).
			Required().
			Unique().
			Field("knowledge_entity_id"),
	}
}

func (SystemAnalysisTopologyNode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "analysis_id", "knowledge_entity_id").Unique(),
	}
}

type SystemAnalysisTopologyEdge struct {
	ent.Schema
}

func (SystemAnalysisTopologyEdge) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SystemAnalysisTopologyEdge) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("analysis_id", uuid.UUID{}),
		field.UUID("knowledge_relationship_id", uuid.UUID{}),
		field.Time("referenced_at").Default(time.Now),
		field.Text("description").Optional(),
	}
}

func (SystemAnalysisTopologyEdge) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("analysis", SystemAnalysis.Type).
			Required().
			Unique().
			Field("analysis_id"),
		edge.To("knowledge_relationship", KnowledgeRelationship.Type).
			Required().
			Unique().
			Field("knowledge_relationship_id"),
	}
}

func (SystemAnalysisTopologyEdge) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "analysis_id", "knowledge_relationship_id").Unique(),
	}
}
