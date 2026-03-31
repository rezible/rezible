package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Alert struct {
	ent.Schema
}

func (Alert) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		IntegrationDataMixin{},
	}
}

func (Alert) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("title"),
		field.String("description").Optional(),
		field.String("definition").Optional(),
		field.UUID("roster_id", uuid.UUID{}).Optional(),
	}
}

// Edges of the Alert.
func (Alert) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("playbooks", Playbook.Type).Ref("alerts"),
		edge.From("roster", OncallRoster.Type).Ref("alerts").Unique().Field("roster_id"),

		edge.To("instances", AlertInstance.Type),
	}
}

type AlertInstance struct {
	ent.Schema
}

func (AlertInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		IntegrationDataMixin{},
	}
}

func (AlertInstance) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("alert_id", uuid.UUID{}),
		field.UUID("event_id", uuid.UUID{}),
		field.Time("acknowledged_at").Optional(),
		//field.UUID("acknowledged_by_id", uuid.UUID{}).Optional(),
	}
}

func (AlertInstance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("alert", Alert.Type).
			Required().
			Unique().
			Field("alert_id"),
		edge.To("event", Event.Type).
			Required().
			Unique().
			Field("event_id"),
		edge.To("feedback", AlertFeedback.Type).Unique(),
	}
}

type AlertFeedback struct {
	ent.Schema
}

func (AlertFeedback) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (AlertFeedback) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("alert_instance_id", uuid.UUID{}),
		field.Bool("actionable"),
		field.Enum("accurate").Values("yes", "no", "unknown"),
		field.Bool("documentation_available"),
		field.Bool("documentation_needs_update"),
	}
}

func (AlertFeedback) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("alert_instance", AlertInstance.Type).
			Required().
			Unique().
			Field("alert_instance_id"),
	}
}
