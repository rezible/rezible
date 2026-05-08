package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type SystemAnalysis struct {
	ent.Schema
}

func (SystemAnalysis) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SystemAnalysis) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("topology_snapshot_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SystemAnalysis) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("retrospective", Retrospective.Type).
			Unique().Required(),
		edge.To("topology_snapshot", SystemTopologySnapshot.Type).
			Unique().
			Field("topology_snapshot_id"),
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
		field.UUID("snapshot_entity_id", uuid.UUID{}),
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
		edge.To("snapshot_entity", SystemTopologySnapshotEntity.Type).
			Required().
			Unique().
			Field("snapshot_entity_id"),
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
		field.UUID("snapshot_relationship_id", uuid.UUID{}),
		field.Text("description").Optional(),
	}
}

func (SystemAnalysisTopologyEdge) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("analysis", SystemAnalysis.Type).
			Required().
			Unique().
			Field("analysis_id"),
		edge.To("snapshot_relationship", SystemTopologySnapshotRelationship.Type).
			Required().
			Unique().
			Field("snapshot_relationship_id"),
	}
}
