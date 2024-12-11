package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// OncallUserShiftCover holds the schema definition for the OncallUserShiftCover entity.
type OncallUserShiftCover struct {
	ent.Schema
}

// Fields of the OncallUserShiftCover.
func (OncallUserShiftCover) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("shift_id", uuid.UUID{}),
		field.Time("start_at"),
		field.Time("end_at"),
		field.String("provider_id").Unique().Optional(),
	}
}

// Edges of the OncallUserShiftCover.
func (OncallUserShiftCover) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().Required().Field("user_id"),
		edge.From("shift", OncallUserShift.Type).
			Ref("covers").
			Unique().Required().Field("shift_id"),
	}
}
