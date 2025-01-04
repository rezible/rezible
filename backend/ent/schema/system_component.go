package schema

/*
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
		field.Int("x_position").
			Optional(), // For diagram layout
		field.Int("y_position").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SystemComponent) Edges() []ent.Edge {
	return []ent.Edge{
		// Incident involvement
		edge.From("incidents", Incident.Type).
			Ref("system_components").
			Through("incident_system_components", IncidentSystemComponent.Type),
		// Hierarchical relationships
		edge.To("children", SystemComponent.Type).
			From("parent").
			Unique(),
		// Control relationships
		edge.To("controls", SystemComponent.Type).
			Through("control_relationships", ControlRelationship.Type),
		// Feedback relationships
		edge.To("feedback_to", SystemComponent.Type).
			Through("feedback_relationships", FeedbackRelationship.Type),
		// Timeline events this component was involved in
		edge.From("events", IncidentEvent.Type).
			Ref("system_components").
			Through("event_components", IncidentEventSystemComponent.Type),
	}
}

type IncidentSystemComponent struct {
	ent.Schema
}

func (IncidentSystemComponent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.UUID("system_component_id", uuid.UUID{}),
		field.Enum("role").
			Values(
				"primary",      // Root cause
				"contributing", // Contributing factor
				"affected",     // Impacted by incident
				"mitigating",   // Used in mitigation
			),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (IncidentSystemComponent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("incident", Incident.Type).Unique().Required().Field("incident_id"),
		edge.To("system_component", SystemComponent.Type).Unique().Required().Field("system_component_id"),
	}
}

type ControlRelationship struct {
	ent.Schema
}

func (ControlRelationship) Fields() []ent.Field {
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

func (ControlRelationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("controller", SystemComponent.Type).Unique().Required().Field("controller_id"),
		edge.To("controlled", SystemComponent.Type).Unique().Required().Field("controlled_id"),
	}
}

type FeedbackRelationship struct {
	ent.Schema
}

func (FeedbackRelationship) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.String("type").
			NotEmpty(), // e.g., "metrics", "alerts", "logs"
		field.Text("description").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (FeedbackRelationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("source", SystemComponent.Type).Unique().Required(),
		edge.To("target", SystemComponent.Type).Unique().Required(),
	}
}
*/
