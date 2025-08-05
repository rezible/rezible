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

func (IncidentSeverity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		ArchiveMixin{},
	}
}

// Fields of the IncidentSeverity.
func (IncidentSeverity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("provider_id").Optional(),
		field.String("name"),
		field.Int("rank"),
		field.String("color").Optional(),
		field.String("description").Optional(),
	}
}

// Edges of the IncidentSeverity.
func (IncidentSeverity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incidents", Incident.Type).Ref("severity"),

		edge.From("debrief_questions", IncidentDebriefQuestion.Type).Ref("incident_severities"),
	}
}
