package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

// IncidentDebriefMessage holds the schema definition for the IncidentDebriefMessage entity.
type IncidentDebriefMessage struct {
	ent.Schema
}

// Fields of the IncidentDebriefMessage.
func (IncidentDebriefMessage) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("debrief_id", uuid.UUID{}),
		field.UUID("question_id", uuid.UUID{}).Optional(),
		field.Time("created_at").Default(time.Now),
		field.Enum("type").Values("user", "assistant", "question"),
		field.Enum("requested_tool").Values("rating").Optional(),
		field.Text("body"),
	}
}

// Edges of the IncidentDebriefMessage.
func (IncidentDebriefMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("debrief", IncidentDebrief.Type).
			Ref("messages").Unique().Required().Field("debrief_id"),
		edge.To("from_question", IncidentDebriefQuestion.Type).Unique().
			Field("question_id"),
	}
}
