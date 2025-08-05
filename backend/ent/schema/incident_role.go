package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// IncidentRole holds the schema definition for the IncidentRole entity.
type IncidentRole struct {
	ent.Schema
}

func (IncidentRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		ArchiveMixin{},
	}
}

// Fields of the IncidentRole.
func (IncidentRole) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name"),
		field.String("provider_id"),
		field.Bool("required").Default(false),
	}
}

// Edges of the IncidentRole.
func (IncidentRole) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("assignments", IncidentRoleAssignment.Type).Ref("role"),
		edge.From("debrief_questions", IncidentDebriefQuestion.Type).Ref("incident_roles"),
	}
}
