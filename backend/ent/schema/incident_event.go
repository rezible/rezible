package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// IncidentEvent holds the schema definition for the IncidentEvent entity.
type IncidentEvent struct {
	ent.Schema
}

var eventTypes = []string{
	"impact_start",
	"impact_mitigated",
	"impact_end",
	"other",
}

// Fields of the IncidentEvent.
func (IncidentEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.Enum("type").Values(eventTypes...),
		field.Time("time"),
		field.UUID("incident_id", uuid.UUID{}),
	}
}

// Edges of the IncidentEvent.
func (IncidentEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incident", Incident.Type).
			Ref("events").Unique().Required().Field("incident_id"),
		edge.To("services", Service.Type),
	}
}
