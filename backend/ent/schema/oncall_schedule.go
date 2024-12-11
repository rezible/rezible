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

// Mixin of the OncallSchedule.
func (OncallSchedule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ArchiveMixin{},
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
