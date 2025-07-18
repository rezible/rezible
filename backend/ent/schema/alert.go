package schema

import (
	"entgo.io/ent/schema/edge"
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Alert holds the schema definition for the Alert entity.
type Alert struct {
	ent.Schema
}

// Fields of the Alert.
func (Alert) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("title"),
		field.String("provider_id"),
	}
}

// Edges of the Alert.
func (Alert) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("playbooks", Playbook.Type).Ref("alerts"),
		edge.From("instances", OncallEvent.Type).Ref("alert"),
	}
}
