package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
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

type OncallUserShiftCover struct {
	ent.Schema
}

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

func (OncallUserShiftCover) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().Required().Field("user_id"),
		edge.From("shift", OncallUserShift.Type).
			Ref("covers").
			Unique().Required().Field("shift_id"),
	}
}

type OncallUserShiftHandover struct {
	ent.Schema
}

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

func (OncallUserShiftHandover) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("shift", OncallUserShift.Type).
			Ref("handover").Unique().Required().Field("shift_id"),
		edge.To("pinned_annotations", OncallAnnotation.Type),
	}
}
