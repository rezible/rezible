package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

type IncidentEventEvidence struct {
	ent.Schema
}

func (IncidentEventEvidence) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.Enum("evidence_type").
			Values("log", "metric", "chat", "ticket", "other"),
		field.String("url").
			NotEmpty(),
		field.String("title").
			NotEmpty(),
		field.Text("description").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (IncidentEventEvidence) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", IncidentEvent.Type).
			Ref("evidence").
			Unique().
			Required(),
	}
}
