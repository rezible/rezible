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
		TenantMixin{},
		ArchiveMixin{},
		IntegrationMixin{},
	}
}

// Fields of the IncidentRole.
func (IncidentRole) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name"),
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

type IncidentRoleAssignment struct {
	ent.Schema
}

func (IncidentRoleAssignment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (IncidentRoleAssignment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("role_id", uuid.UUID{}),
	}
}

func (IncidentRoleAssignment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("incident", Incident.Type).
			Required().
			Unique().
			Field("incident_id"),
		edge.To("user", User.Type).
			Required().
			Unique().
			Field("user_id"),
		edge.To("role", IncidentRole.Type).
			Required().
			Unique().
			Field("role_id"),
	}
}
