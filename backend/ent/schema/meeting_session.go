package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

// MeetingSession holds the schema definition for the MeetingSession entity.
type MeetingSession struct {
	ent.Schema
}

func (MeetingSession) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields of the MeetingSession.
func (MeetingSession) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("title"),
		field.Time("started_at").Default(time.Now),
		field.Time("ended_at").Optional(),
		field.String("document_name"),
	}
}

// Edges of the MeetingSession.
func (MeetingSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incidents", Incident.Type).
			Ref("review_sessions"),
	}
}
