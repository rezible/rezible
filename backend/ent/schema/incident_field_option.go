package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

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
