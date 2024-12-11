package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// IncidentSeverity holds the schema definition for the IncidentSeverity entity.
type IncidentSeverity struct {
	ent.Schema
}

// Fields of the IncidentSeverity.
func (IncidentSeverity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("name"),
		field.String("color").Optional(),
		field.String("description").Optional(),
	}
}

// Mixin of the IncidentSeverity.
func (IncidentSeverity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ArchiveMixin{},
	}
}

// Edges of the IncidentSeverity.
func (IncidentSeverity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incidents", Incident.Type).Ref("severity"),

		edge.From("debrief_questions", IncidentDebriefQuestion.Type).Ref("incident_severities"),
	}
}
