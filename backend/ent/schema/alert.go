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

func (Alert) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Edges of the Alert.
func (Alert) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("metrics", AlertMetrics.Type).Ref("alert"),
		edge.From("playbooks", Playbook.Type).Ref("alerts"),
		edge.From("instances", OncallEvent.Type).Ref("alert"),
	}
}

// AlertMetrics holds the schema definition for the AlertMetrics entity.
type AlertMetrics struct {
	ent.Schema
}

// Fields of the AlertMetrics.
func (AlertMetrics) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("alert_id", uuid.UUID{}),
	}
}

func (AlertMetrics) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Edges of the AlertMetrics.
func (AlertMetrics) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("alert", Alert.Type).Unique().Required().Field("alert_id"),
	}
}
