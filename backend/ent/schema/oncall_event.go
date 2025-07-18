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

var oncallEventKind = []string{"alert", "page", "message", "other"}

func (OncallEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("provider_id"),
		field.UUID("roster_id", uuid.UUID{}).Optional(),
		field.Time("timestamp"),
		field.Enum("kind").Values(oncallEventKind...),
		field.String("title"),
		field.String("description"),
		field.String("source"),
	}
}

func (OncallEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("roster", OncallRoster.Type).Field("roster_id").Unique(),
		edge.To("alert", Alert.Type).Unique(),
		edge.From("annotations", OncallAnnotation.Type).Ref("event"),
	}
}
