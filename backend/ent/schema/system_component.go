package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

type SystemComponent struct {
	ent.Schema
}

func (SystemComponent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.String("name").
			NotEmpty(),
		field.Enum("type").
			Values(
				"service",          // e.g., API service, database
				"control",          // e.g., rate limiter, circuit breaker
				"feedback",         // e.g., monitoring, alerts
				"interface",        // e.g., API endpoint, UI
				"human_controller", // e.g., team, operations
			),
		field.Text("description").
			Optional(),
		field.JSON("properties", map[string]any{}), // Flexible properties based on component type
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SystemComponent) Edges() []ent.Edge {
	return []ent.Edge{
		// Hierarchical relationships
		//edge.To("children", SystemComponent.Type).
		//	From("parent").
		//	Unique(),
		edge.To("analyses", SystemAnalysis.Type).
			Through("analysis_components", SystemAnalysisComponent.Type),
		// relationships
		edge.To("related", SystemComponent.Type).
			StorageKey(edge.Table("system_component_relationship"), edge.Columns("source_id", "target_id")).
			Through("component_relationships", SystemComponentRelationship.Type),
		//edge.To("controls", SystemComponent.Type).
		//	StorageKey(edge.Table("system_component_control_relationship"), edge.Columns("controller_id", "controlled_id")).
		//	Through("control_relationships", SystemComponentControlRelationship.Type),
		//// Feedback relationships
		//edge.To("feedback_to", SystemComponent.Type).
		//	StorageKey(edge.Table("system_component_feedback_relationship"), edge.Columns("source_id", "target_id")).
		//	Through("feedback_relationships", SystemComponentFeedbackRelationship.Type),
		// Incident involvement
		//edge.From("incidents", Incident.Type).
		//	Ref("system_components").
		//	Through("incident_system_components", IncidentSystemComponent.Type),
		// Timeline events this component was involved in
		edge.From("events", IncidentEvent.Type).
			Ref("system_components").
			Through("event_components", IncidentEventSystemComponent.Type),

		// TODO: constraints
		edge.From("constraints", SystemComponentConstraint.Type).
			Ref("component"),
		edge.From("controls", SystemComponentControl.Type).
			Ref("component"),
		edge.From("signals", SystemComponentSignal.Type).
			Ref("component"),
	}
}

type SystemComponentConstraint struct {
	ent.Schema
}

func (SystemComponentConstraint) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("component_id", uuid.UUID{}),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemComponentConstraint) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("component", SystemComponent.Type).
			Required().Unique().Field("component_id"),
	}
}

type SystemComponentSignal struct {
	ent.Schema
}

func (SystemComponentSignal) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("component_id", uuid.UUID{}),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemComponentSignal) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("component", SystemComponent.Type).
			Required().Unique().Field("component_id"),
		edge.From("feedback_signals", SystemComponentRelationshipFeedback.Type).
			Ref("signal"),
	}
}

type SystemComponentControl struct {
	ent.Schema
}

func (SystemComponentControl) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("component_id", uuid.UUID{}),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemComponentControl) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("component", SystemComponent.Type).
			Required().Unique().Field("component_id"),
		edge.From("control_actions", SystemComponentRelationshipControlAction.Type).
			Ref("control"),
	}
}

type SystemComponentRelationship struct {
	ent.Schema
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
			Required().Unique().Field("source_id"),
		edge.To("target", SystemComponent.Type).
			Required().Unique().Field("target_id"),
		edge.To("control_actions", SystemComponentRelationshipControlAction.Type),
		edge.To("feedback", SystemComponentRelationshipFeedback.Type),
	}
}

type SystemComponentRelationshipControlAction struct {
	ent.Schema
}

func (SystemComponentRelationshipControlAction) Fields() []ent.Field {
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

func (SystemComponentRelationshipControlAction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("control", SystemComponentControl.Type).
			Unique().Required().Field("control_id"),
		edge.From("relationship", SystemComponentRelationship.Type).
			Ref("control_actions").
			Unique().Required().Field("relationship_id"),
	}
}

type SystemComponentRelationshipFeedback struct {
	ent.Schema
}

func (SystemComponentRelationshipFeedback) Fields() []ent.Field {
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

func (SystemComponentRelationshipFeedback) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("signal", SystemComponentSignal.Type).
			Unique().Required().Field("signal_id"),
		edge.From("relationship", SystemComponentRelationship.Type).
			Ref("feedback").
			Unique().Required().Field("relationship_id"),
	}
}
