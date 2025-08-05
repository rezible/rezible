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

func (OncallUserShift) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields of the OncallUserShift.
func (OncallUserShift) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("roster_id", uuid.UUID{}),
		field.String("provider_id").Unique().Optional(),
		field.Enum("role").Values("primary", "secondary", "shadow", "covering").Default("primary").Optional(),
		field.UUID("primary_shift_id", uuid.UUID{}).Optional(),
		field.Time("start_at"),
		field.Time("end_at"),
	}
}

// Edges of the OncallUserShift.
func (OncallUserShift) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().Required().Field("user_id"),
		edge.To("roster", OncallRoster.Type).
			Unique().Required().Field("roster_id"),
		edge.To("primary_shift", OncallUserShift.Type).Field("primary_shift_id").Unique(),

		edge.To("handover", OncallUserShiftHandover.Type).Unique(),
		edge.To("metrics", OncallUserShiftMetrics.Type).Unique(),
	}
}

type OncallUserShiftHandover struct {
	ent.Schema
}

func (OncallUserShiftHandover) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
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

type OncallUserShiftMetrics struct {
	ent.Schema
}

func (OncallUserShiftMetrics) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (OncallUserShiftMetrics) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("shift_id", uuid.UUID{}),
	}
}

func (OncallUserShiftMetrics) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("shift", OncallUserShift.Type).Ref("metrics").Unique().Required().Field("shift_id"),
	}
}
