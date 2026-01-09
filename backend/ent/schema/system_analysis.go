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
		edge.From("components", SystemComponent.Type).
			Ref("system_analyses").
			Through("analysis_components", SystemAnalysisComponent.Type),
		edge.From("relationships", SystemAnalysisRelationship.Type).
			Ref("system_analysis"),
	}
}

type SystemAnalysisComponent struct {
	ent.Schema
}

func (SystemAnalysisComponent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SystemAnalysisComponent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("analysis_id", uuid.UUID{}),
		field.UUID("component_id", uuid.UUID{}),
		field.Text("description").Optional(),
		field.Float("pos_x").Default(0),
		field.Float("pos_y").Default(0),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemAnalysisComponent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("analysis", SystemAnalysis.Type).
			Required().Unique().Field("analysis_id"),
		edge.To("component", SystemComponent.Type).
			Required().Unique().Field("component_id"),
	}
}

type SystemAnalysisRelationship struct {
	ent.Schema
}

func (SystemAnalysisRelationship) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SystemAnalysisRelationship) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("analysis_id", uuid.UUID{}),
		field.UUID("component_relationship_id", uuid.UUID{}),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemAnalysisRelationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("system_analysis", SystemAnalysis.Type).
			Required().
			Unique().
			Field("analysis_id"),

		edge.To("component_relationship", SystemComponentRelationship.Type).
			Unique().
			Required().
			Field("component_relationship_id"),
	}
}
