package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// IncidentMilestone holds the schema definition for the IncidentMilestone entity.
type IncidentMilestone struct {
	ent.Schema
}

// Fields of the IncidentMilestone.
func (IncidentMilestone) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.Enum("kind").Values("impact", "detected", "investigating", "mitigated", "resolved"),
		field.Time("time"),
	}
}

// Edges of the IncidentEvent.
func (IncidentMilestone) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incident", Incident.Type).
			Ref("milestones").Unique().Required().Field("incident_id"),
	}
}
