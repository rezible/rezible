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

// Fields of the IncidentType.
func (IncidentType) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("name"),
	}
}

// Mixin of the IncidentType.
func (IncidentType) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ArchiveMixin{},
	}
}

// Edges of the IncidentType.
func (IncidentType) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incidents", Incident.Type).Ref("type"),

		edge.From("debrief_questions", IncidentDebriefQuestion.Type).Ref("incident_types"),
	}
}
