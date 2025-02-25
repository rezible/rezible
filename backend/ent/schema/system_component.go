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
		field.UUID("kind_id", uuid.UUID{}),
		//field.Enum("type").
		//	Values(
		//		"service",          // e.g., API service, database
		//		"control",          // e.g., rate limiter, circuit breaker
		//		"feedback",         // e.g., monitoring, alerts
		//		"interface",        // e.g., API endpoint, UI
		//		"human_controller", // e.g., team, operations
		//	),
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
		edge.To("kind", SystemComponentKind.Type).
			Required().Unique().
			Field("kind_id"),
		edge.To("analyses", SystemAnalysis.Type).
			Through("analysis_components", SystemAnalysisComponent.Type),
		// relationships
		edge.To("related", SystemComponent.Type).
			StorageKey(
				edge.Table("system_relationship"),
				edge.Columns("source_component_id", "target_component_id")).
			Through("relationships", SystemAnalysisRelationship.Type),
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

		edge.From("constraints", SystemComponentConstraint.Type).
			Ref("component"),
		edge.From("controls", SystemComponentControl.Type).
			Ref("component"),
		edge.From("signals", SystemComponentSignal.Type).
			Ref("component"),
	}
}

type SystemComponentKind struct {
	ent.Schema
}

func (SystemComponentKind) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("label"),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemComponentKind) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("components", SystemComponent.Type).Ref("kind"),
	}
}

type SystemComponentConstraint struct {
	ent.Schema
}

func (SystemComponentConstraint) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("component_id", uuid.UUID{}),
		field.Text("label"),
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
		field.Text("label"),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemComponentSignal) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("component", SystemComponent.Type).
			Required().Unique().Field("component_id"),
		edge.From("relationships", SystemAnalysisRelationship.Type).
			Ref("signals").
			Through("feedback_signals", SystemRelationshipFeedbackSignal.Type),
	}
}

type SystemComponentControl struct {
	ent.Schema
}

func (SystemComponentControl) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("component_id", uuid.UUID{}),
		field.Text("label"),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemComponentControl) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("component", SystemComponent.Type).
			Required().Unique().Field("component_id"),
		edge.From("relationships", SystemAnalysisRelationship.Type).
			Ref("controls").
			Through("control_actions", SystemRelationshipControlAction.Type),
	}
}
