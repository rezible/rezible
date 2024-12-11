package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// IncidentRoleAssignment holds the schema definition for the IncidentRoleAssignment entity.
type IncidentRoleAssignment struct {
	ent.Schema
}

// Fields of the IncidentRoleAssignment.
func (IncidentRoleAssignment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("role_id", uuid.UUID{}),
		field.UUID("incident_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
	}
}

func (IncidentRoleAssignment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// field.ID("user_id", "incident_id", "role"),
	}
}

// Edges of the IncidentRoleAssignment.
func (IncidentRoleAssignment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("role", IncidentRole.Type).
			Unique().Required().Field("role_id"),
		edge.To("incident", Incident.Type).
			Unique().Required().Field("incident_id"),
		edge.To("user", User.Type).
			Unique().Required().Field("user_id"),
	}
}
