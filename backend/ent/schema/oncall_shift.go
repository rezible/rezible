package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type OncallShift struct {
	ent.Schema
}

func (OncallShift) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

// Fields of the OncallShift.
func (OncallShift) Fields() []ent.Field {
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

// Edges of the OncallShift.
func (OncallShift) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required().
			Field("user_id"),

		edge.To("roster", OncallRoster.Type).
			Unique().
			Required().
			Field("roster_id"),

		edge.To("primary_shift", OncallShift.Type).
			Unique().
			Field("primary_shift_id"),

		edge.To("handover", OncallShiftHandover.Type).Unique(),
		edge.To("metrics", OncallShiftMetrics.Type).Unique(),
	}
}

type OncallShiftHandover struct {
	ent.Schema
}

func (OncallShiftHandover) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (OncallShiftHandover) Fields() []ent.Field {
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

func (OncallShiftHandover) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("shift", OncallShift.Type).
			Ref("handover").
			Unique().
			Required().
			Field("shift_id"),
		edge.To("pinned_annotations", OncallAnnotation.Type),
	}
}

type OncallShiftMetrics struct {
	ent.Schema
}

func (OncallShiftMetrics) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (OncallShiftMetrics) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("shift_id", uuid.UUID{}),
		field.Time("updated_at").Default(time.Now),
		// Burden
		field.Float32("burden_score"),
		field.Float32("event_frequency"),
		field.Float32("life_impact"),
		field.Float32("time_impact"),
		field.Float32("response_requirements"),
		field.Float32("isolation"),
		// Incidents
		field.Float32("incidents_total"),
		field.Float32("incident_response_time"),
		// Interrupts
		field.Float32("interrupts_total"),
		field.Float32("interrupts_alerts"),
		field.Float32("interrupts_night"),
		field.Float32("interrupts_business_hours"),
	}
}

func (OncallShiftMetrics) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("shift", OncallShift.Type).
			Ref("metrics").
			Unique().
			Required().
			Field("shift_id"),
	}
}
