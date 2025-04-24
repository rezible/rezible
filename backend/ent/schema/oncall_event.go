package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type OncallEvent struct {
	ent.Schema
}

func (OncallEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("provider_id"),
		field.Time("timestamp"),
		field.String("kind"),
		field.String("title"),
		field.String("description"),
		field.String("source"),
	}
}

func (OncallEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("annotations", OncallAnnotation.Type).Ref("event"),
	}
}
