package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

type SystemAnalysisRelationship struct {
	ent.Schema
}

func (SystemAnalysisRelationship) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("analysis_id", uuid.UUID{}),
		field.UUID("source_component_id", uuid.UUID{}),
		field.UUID("target_component_id", uuid.UUID{}),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemAnalysisRelationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("system_analysis", SystemAnalysis.Type).
			Required().Unique().Field("analysis_id"),

		edge.To("source_component", SystemComponent.Type).
			Required().Unique().Field("source_component_id"),
		edge.To("target_component", SystemComponent.Type).
			Required().Unique().Field("target_component_id"),

		edge.To("controls", SystemComponentControl.Type).
			Through("control_actions", SystemRelationshipControlAction.Type),
		edge.To("signals", SystemComponentSignal.Type).
			Through("feedback_signals", SystemRelationshipFeedbackSignal.Type),
	}
}

type SystemRelationshipControlAction struct {
	ent.Schema
}

func (SystemRelationshipControlAction) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("relationship_id", uuid.UUID{}),
		field.UUID("control_id", uuid.UUID{}),
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
		edge.To("relationship", SystemAnalysisRelationship.Type).
			Unique().Required().Field("relationship_id"),
		edge.To("control", SystemComponentControl.Type).
			Unique().Required().Field("control_id"),
	}
}

type SystemRelationshipFeedbackSignal struct {
	ent.Schema
}

func (SystemRelationshipFeedbackSignal) Fields() []ent.Field {
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

func (SystemRelationshipFeedbackSignal) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("relationship", SystemAnalysisRelationship.Type).
			Unique().Required().Field("relationship_id"),
		edge.To("signal", SystemComponentSignal.Type).
			Unique().Required().Field("signal_id"),
	}
}
