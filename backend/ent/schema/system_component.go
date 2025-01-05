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
				"human_controller", // e.g., SRE team, operations
			),
		field.Text("description").
			Optional(),
		field.JSON("properties", map[string]any{}).
			Optional(), // Flexible properties based on component type
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SystemComponent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", SystemComponent.Type).
			From("parent").
			Unique(),
		// Control relationships
		edge.To("controls", SystemComponent.Type).
			StorageKey(edge.Table("system_component_control_relationship"), edge.Columns("controller_id", "controlled_id")).
			Through("control_relationships", SystemComponentControlRelationship.Type),
		// Feedback relationships
		edge.To("feedback_to", SystemComponent.Type).
			StorageKey(edge.Table("system_component_feedback_relationship"), edge.Columns("source_id", "target_id")).
			Through("feedback_relationships", SystemComponentFeedbackRelationship.Type),
		// Incident involvement
		edge.From("incidents", Incident.Type).
			Ref("system_components").
			Through("incident_system_components", IncidentSystemComponent.Type),
		// Hierarchical relationships
		// Timeline events this component was involved in
		edge.From("events", IncidentEvent.Type).
			Ref("system_components").
			Through("event_components", IncidentEventSystemComponent.Type),
	}
}

type SystemComponentControlRelationship struct {
	ent.Schema
}

func (SystemComponentControlRelationship) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("controller_id", uuid.UUID{}),
		field.UUID("controlled_id", uuid.UUID{}),
		field.String("type").
			NotEmpty(), // e.g., "rate_limits", "circuit_breaks"
		field.Text("description").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (SystemComponentControlRelationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("controller", SystemComponent.Type).
			Required().Unique().Field("controller_id"),
		edge.To("controlled", SystemComponent.Type).
			Required().Unique().Field("controlled_id"),
	}
}

type SystemComponentFeedbackRelationship struct {
	ent.Schema
}

func (SystemComponentFeedbackRelationship) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("source_id", uuid.UUID{}),
		field.UUID("target_id", uuid.UUID{}),
		field.String("type").
			NotEmpty(), // e.g., "metrics", "alerts", "logs"
		field.Text("description").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (SystemComponentFeedbackRelationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("source", SystemComponent.Type).
			Unique().Required().Field("source_id"),
		edge.To("target", SystemComponent.Type).
			Unique().Required().Field("target_id"),
	}
}
