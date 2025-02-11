package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

type SystemRelationship struct {
	ent.Schema
}

func (SystemRelationship) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("source_component_id", uuid.UUID{}),
		field.UUID("target_component_id", uuid.UUID{}),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemRelationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("source_component", SystemComponent.Type).
			Required().Unique().Field("source_component_id"),
		edge.To("target_component", SystemComponent.Type).
			Required().Unique().Field("target_component_id"),

		edge.To("controls", SystemComponentControl.Type).
			Through("control_actions", SystemRelationshipControlAction.Type),
		edge.To("signals", SystemComponentSignal.Type).
			Through("feedback", SystemRelationshipFeedback.Type),
	}
}

type SystemRelationshipControlAction struct {
	ent.Schema
}

func (SystemRelationshipControlAction) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("control_id", uuid.UUID{}),
		field.UUID("relationship_id", uuid.UUID{}),
		field.String("type").
			NotEmpty(), // e.g., "rate_limits", "circuit_breaks"
		field.Text("description").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (SystemRelationshipControlAction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("control", SystemComponentControl.Type).
			Unique().Required().Field("control_id"),
		edge.To("relationship", SystemRelationship.Type).
			Unique().Required().Field("relationship_id"),
	}
}

type SystemRelationshipFeedback struct {
	ent.Schema
}

func (SystemRelationshipFeedback) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("relationship_id", uuid.UUID{}),
		field.UUID("signal_id", uuid.UUID{}),
		field.String("type").
			NotEmpty(), // e.g., "metrics", "alerts", "logs"
		field.Text("description").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (SystemRelationshipFeedback) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("signal", SystemComponentSignal.Type).
			Unique().Required().Field("signal_id"),
		edge.To("relationship", SystemRelationship.Type).
			Unique().Required().Field("relationship_id"),
	}
}
