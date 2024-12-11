package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// OncallScheduleParticipant holds the schema definition for the OncallScheduleParticipant entity.
type OncallScheduleParticipant struct {
	ent.Schema
}

// Fields of the OncallScheduleParticipant.
func (OncallScheduleParticipant) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("schedule_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Int("index"),
	}
}

func (OncallScheduleParticipant) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// field.ID("user_id", "incident_id", "role"),
	}
}

// Edges of the OncallScheduleParticipant.
func (OncallScheduleParticipant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("schedule", OncallSchedule.Type).
			Ref("participants").
			Unique().Required().Field("schedule_id"),
		edge.To("user", User.Type).
			Unique().Required().Field("user_id"),
	}
}
