package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// OncallUserShift holds the schema definition for the OncallUserShift entity.
type OncallUserShift struct {
	ent.Schema
}

// Fields of the OncallUserShift.
func (OncallUserShift) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("roster_id", uuid.UUID{}),
		field.Time("start_at"),
		field.Time("end_at"),
		field.String("provider_id").Unique().Optional(),
	}
}

// Edges of the OncallUserShift.
func (OncallUserShift) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().Required().Field("user_id"),
		edge.To("roster", OncallRoster.Type).
			Unique().Required().Field("roster_id"),

		edge.To("covers", OncallUserShiftCover.Type),
		edge.To("handover", OncallUserShiftHandover.Type).Unique(),
	}
}
