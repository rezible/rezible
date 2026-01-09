package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type SystemComponentRelationship struct {
	ent.Schema
}

func (SystemComponentRelationship) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		IntegrationMixin{},
	}
}

func (SystemComponentRelationship) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("source_id", uuid.UUID{}),
		field.UUID("target_id", uuid.UUID{}),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemComponentRelationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("source", SystemComponent.Type).
			Required().Unique().
			Field("source_id"),
		edge.To("target", SystemComponent.Type).
			Required().Unique().
			Field("target_id"),

		edge.From("system_analyses", SystemAnalysisRelationship.Type).
			Ref("component_relationship"),

		edge.From("hazards", SystemHazard.Type).
			Ref("relationships"),

		edge.To("controls", SystemComponentControl.Type).
			Through("control_actions", SystemRelationshipControlAction.Type),
		edge.To("signals", SystemComponentSignal.Type).
			Through("feedback_signals", SystemRelationshipFeedbackSignal.Type),
	}
}

type SystemRelationshipControlAction struct {
	ent.Schema
}

func (SystemRelationshipControlAction) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SystemRelationshipControlAction) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("relationship_id", uuid.UUID{}),
		field.UUID("control_id", uuid.UUID{}),
		field.String("type").
			NotEmpty(),
		field.Text("description").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (SystemRelationshipControlAction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("relationship", SystemComponentRelationship.Type).
			Unique().Required().Field("relationship_id"),
		edge.To("control", SystemComponentControl.Type).
			Unique().Required().Field("control_id"),
	}
}

type SystemRelationshipFeedbackSignal struct {
	ent.Schema
}

func (SystemRelationshipFeedbackSignal) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SystemRelationshipFeedbackSignal) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("relationship_id", uuid.UUID{}),
		field.UUID("signal_id", uuid.UUID{}),
		field.String("type").
			NotEmpty(),
		field.Text("description").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (SystemRelationshipFeedbackSignal) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("relationship", SystemComponentRelationship.Type).
			Unique().Required().Field("relationship_id"),
		edge.To("signal", SystemComponentSignal.Type).
			Unique().Required().Field("signal_id"),
	}
}
