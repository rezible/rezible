package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// IncidentTag holds the schema definition for the IncidentTag entity.
type IncidentTag struct {
	ent.Schema
}

// Fields of the IncidentTag.
func (IncidentTag) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("key"),
		field.String("value"),
	}
}

// Mixin of the IncidentTag.
func (IncidentTag) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ArchiveMixin{},
	}
}

// Edges of the IncidentTag.
func (IncidentTag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incidents", Incident.Type).Ref("tag_assignments"),
		edge.From("debrief_questions", IncidentDebriefQuestion.Type).Ref("incident_tags"),
	}
}
