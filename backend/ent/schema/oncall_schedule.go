package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// OncallSchedule holds the schema definition for the OncallSchedule entity.
type OncallSchedule struct {
	ent.Schema
}

func (OncallSchedule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		ArchiveMixin{},
	}
}

// Fields of the OncallSchedule.
func (OncallSchedule) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name"),
		field.UUID("roster_id", uuid.UUID{}),
		field.String("timezone").Optional(),
		field.String("provider_id").Unique(),
		// start, end, cadence, etc
	}
}

// Edges of the OncallSchedule.
func (OncallSchedule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("participants", OncallScheduleParticipant.Type),
		edge.From("roster", OncallRoster.Type).Ref("schedules").
			Unique().Required().Field("roster_id"),
	}
}

type OncallScheduleParticipant struct {
	ent.Schema
}

func (OncallScheduleParticipant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (OncallScheduleParticipant) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("schedule_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Int("index"),
	}
}

func (OncallScheduleParticipant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("schedule", OncallSchedule.Type).
			Ref("participants").Unique().Required().Field("schedule_id"),
		edge.To("user", User.Type).
			Unique().Required().Field("user_id"),
	}
}
