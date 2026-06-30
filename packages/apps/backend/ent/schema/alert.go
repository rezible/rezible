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
		KnowledgeEntityLinkMixin{},
	}
}

func (Alert) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("title"),
		field.String("description").Optional(),
		field.String("definition").Optional(),
	}
}

// Edges of the Alert.
func (Alert) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("playbooks", Playbook.Type).Ref("alerts"),
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
		KnowledgeEntityLinkMixin{},
	}
}

func (AlertInstance) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("alert_id", uuid.UUID{}),
	}
}

func (AlertInstance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("alert", Alert.Type).
			Required().Unique().
			Field("alert_id").
			Ref("instances"),
		edge.From("feedback", AlertFeedback.Type).
			Ref("alert_instance"),
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
			Unique().
			Required().
			Field("alert_instance_id"),
	}
}

type AlertMetrics struct {
	ent.View
}

func (AlertMetrics) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (AlertMetrics) Fields() []ent.Field {
	return []ent.Field{
		field.Int("event_count"),
		field.Int("interrupt_count"),
		field.Int("night_interrupt_count"),
		field.Int("incidents"),
		field.Int("feedback_count"),
		field.Int("feedback_actionable"),
		field.Int("feedback_accurate"),
		field.Int("feedback_accurate_unknown"),
		field.Int("feedback_docs_available"),
		field.Int("feedback_docs_need_update"),
	}
}
