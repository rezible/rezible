package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// IncidentLink holds the schema definition for the IncidentLink entity.
type IncidentLink struct {
	ent.Schema
}

func (IncidentLink) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields of the IncidentLink.
func (IncidentLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("incident_id", uuid.UUID{}),
		field.UUID("linked_incident_id", uuid.UUID{}),
		field.String("description").Optional(),
		field.Enum("link_type").Values("parent", "child", "similar"),
	}
}

// Edges of the IncidentLink.
func (IncidentLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("incident", Incident.Type).
			Required().
			Unique().
			Field("incident_id"),
		edge.To("linked_incident", Incident.Type).
			Required().
			Unique().
			Field("linked_incident_id"),
	}
}
