package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// IncidentType holds the schema definition for the IncidentType entity.
type IncidentType struct {
	ent.Schema
}

func (IncidentType) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		ArchiveMixin{},
	}
}

// Fields of the IncidentType.
func (IncidentType) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("name"),
	}
}

// Edges of the IncidentType.
func (IncidentType) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incidents", Incident.Type).Ref("type"),
		edge.From("debrief_questions", IncidentDebriefQuestion.Type).Ref("incident_types"),
	}
}
