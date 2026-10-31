package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// MeetingSession holds the schema definition for the MeetingSession entity.
type MeetingSession struct {
	ent.Schema
}

func (MeetingSession) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
	}
}

// Fields of the MeetingSession.
func (MeetingSession) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.Time("started_at").Default(time.Now),
		field.Time("ended_at").Optional(),
		field.String("document_name"),
	}
}

// Edges of the MeetingSession.
func (MeetingSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("video_conference", VideoConference.Type).Unique(),
		edge.To("schedule", MeetingSchedule.Type).Unique(),
	}
}
