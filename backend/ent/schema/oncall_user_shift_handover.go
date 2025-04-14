package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

// OncallUserShiftHandover holds the schema definition for the OncallUserShiftHandover entity.
type OncallUserShiftHandover struct {
	ent.Schema
}

// Fields of the OncallUserShiftHandover.
func (OncallUserShiftHandover) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("shift_id", uuid.UUID{}),
		field.Time("created_at"),
		field.Bool("reminder_sent").Default(false),
		field.Time("updated_at").Default(time.Now),
		field.Time("sent_at").Optional(),
		field.Bytes("contents"),
	}
}

// Edges of the OncallUserShiftHandover.
func (OncallUserShiftHandover) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("shift", OncallUserShift.Type).
			Ref("handover").Unique().Required().Field("shift_id"),
		edge.To("pinned_annotations", OncallAnnotation.Type),
	}
}
