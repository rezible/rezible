package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// IncidentEvent holds the schema definition for the IncidentEvent entity.
type IncidentEvent struct {
	ent.Schema
}

// Fields of the IncidentEvent.
func (IncidentEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.Time("timestamp"),
		field.Enum("kind").
			Values("observation", "action", "decision", "context"),
		field.String("title").
			NotEmpty(),
		field.Text("description").
			Optional(),
		field.Bool("is_key").Default(false),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.UUID("created_by", uuid.UUID{}),
		field.Int("sequence").Default(0),
		field.Bool("is_draft").Default(false),
	}
}

// Indexes of the IncidentEvent.
func (IncidentEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("incident_id", "timestamp", "sequence").Unique(),
	}
}

// Edges of the IncidentEvent.
func (IncidentEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incident", Incident.Type).
			Ref("events").
			Field("incident_id").
			Unique().
			Required(),
		edge.To("context", IncidentEventContext.Type).
			Unique(),
		edge.To("factors", IncidentEventContributingFactor.Type),
		edge.To("evidence", IncidentEventEvidence.Type),
		edge.To("system_components", SystemComponent.Type).
			Through("event_components", IncidentEventSystemComponent.Type),
	}
}

type IncidentEventSystemComponent struct {
	ent.Schema
}

func (IncidentEventSystemComponent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("incident_event_id", uuid.UUID{}),
		field.UUID("system_component_id", uuid.UUID{}),
		field.Enum("relationship").
			Values("primary", "affected", "contributing"),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (IncidentEventSystemComponent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", IncidentEventSystemComponent.Type).Unique().Required().Field("incident_event_id"),
		edge.To("system_component", SystemComponent.Type).Unique().Required().Field("system_component_id"),
	}
}
