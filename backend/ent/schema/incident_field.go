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

// IncidentFieldOption holds the schema definition for the IncidentFieldOption entity.
type IncidentFieldOption struct {
	ent.Schema
}

// Fields of the IncidentFieldOption.
func (IncidentFieldOption) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("incident_field_id", uuid.UUID{}),
		field.Enum("type").Values("custom", "derived"),
		field.String("value"),
	}
}

// Mixin of the IncidentFieldOption.
func (IncidentFieldOption) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ArchiveMixin{},
	}
}

// Edges of the IncidentFieldOption.
func (IncidentFieldOption) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incident_field", IncidentField.Type).
			Field("incident_field_id").
			Ref("options").Unique().Required(),
		edge.From("incidents", Incident.Type).Ref("field_selections"),
	}
}
