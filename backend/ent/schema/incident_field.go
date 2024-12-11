package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// IncidentField holds the schema definition for the IncidentField entity.
type IncidentField struct {
	ent.Schema
}

// Fields of the IncidentField.
func (IncidentField) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("name"),
	}
}

// Mixin of the IncidentField.
func (IncidentField) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ArchiveMixin{},
	}
}

// Edges of the IncidentField.
func (IncidentField) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("options", IncidentFieldOption.Type),

		edge.From("debrief_questions", IncidentDebriefQuestion.Type).Ref("incident_fields"),
	}
}
