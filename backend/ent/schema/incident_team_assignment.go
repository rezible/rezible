package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// IncidentTeamAssignment holds the schema definition for the IncidentTeamAssignment entity.
type IncidentTeamAssignment struct {
	ent.Schema
}

// Fields of the IncidentTeamAssignment.
func (IncidentTeamAssignment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("incident_id", uuid.UUID{}),
		field.UUID("team_id", uuid.UUID{}),
	}
}

// Edges of the IncidentTeamAssignment.
func (IncidentTeamAssignment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("incident", Incident.Type).
			Unique().Required().Field("incident_id"),
		edge.To("team", Team.Type).
			Unique().Required().Field("team_id"),
	}
}
