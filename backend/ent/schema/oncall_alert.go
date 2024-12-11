package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// OncallAlert holds the schema definition for the OncallAlert entity.
type OncallAlert struct {
	ent.Schema
}

// Fields of the OncallAlert.
func (OncallAlert) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("roster_id", uuid.UUID{}),
		field.String("name"),
		field.Time("timestamp"),
	}
}

// Edges of the OncallAlert.
func (OncallAlert) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("instances", OncallAlertInstance.Type),
		edge.To("roster", OncallRoster.Type).
			Unique().Required().Field("roster_id"),
	}
}
