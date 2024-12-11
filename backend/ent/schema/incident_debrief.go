package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// IncidentDebrief holds the schema definition for the IncidentDebrief entity.
type IncidentDebrief struct {
	ent.Schema
}

// Fields of the IncidentDebrief.
func (IncidentDebrief) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Bool("required"),
		field.Bool("started"),
	}
}

// Edges of the IncidentDebrief.
func (IncidentDebrief) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incident", Incident.Type).
			Ref("debriefs").Unique().Required().Field("incident_id"),
		edge.From("user", User.Type).
			Ref("incident_debriefs").Unique().Required().Field("user_id"),
		edge.To("messages", IncidentDebriefMessage.Type),
		edge.To("suggestions", IncidentDebriefSuggestion.Type),
	}
}
