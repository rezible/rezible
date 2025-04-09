package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

// OncallUserShiftAnnotation holds the schema definition for the OncallUserShiftAnnotation entity.
type OncallUserShiftAnnotation struct {
	ent.Schema
}

// Fields of the OncallUserShiftAnnotation.
func (OncallUserShiftAnnotation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("shift_id", uuid.UUID{}),
		field.String("event_id"),
		field.Int("minutes_occupied"),
		field.Text("notes"),
		field.Bool("pinned"),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the OncallUserShiftAnnotation.
func (OncallUserShiftAnnotation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("shift", OncallUserShift.Type).
			Ref("annotations").Unique().Required().Field("shift_id"),
	}
}
