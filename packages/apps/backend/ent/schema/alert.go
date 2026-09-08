package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type AlertDefinition struct {
	ent.Schema
}

func (AlertDefinition) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		KnowledgeEntityLinkMixin{},
	}
}

func (AlertDefinition) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("title"),
		field.String("description").Optional(),
		field.String("definition").Optional(),
	}
}

func (AlertDefinition) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("playbooks", Playbook.Type).Ref("alert_definitions"),
		edge.To("episodes", AlertEpisode.Type),
	}
}

type AlertInstance struct {
	ent.Schema
}

func (AlertInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (AlertInstance) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("alert_episode_id", uuid.UUID{}).Immutable(),
		field.UUID("normalized_event_id", uuid.UUID{}).Immutable(),
	}
}

func (AlertInstance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("episode", AlertEpisode.Type).
			Required().
			Unique().
			Immutable().
			Field("alert_episode_id").
			Ref("instances"),
		edge.To("event", NormalizedEvent.Type).
			Required().
			Unique().
			Immutable().
			Field("normalized_event_id"),
		edge.From("feedback", AlertFeedback.Type).
			Ref("alert_instance"),
	}
}

func (AlertInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "normalized_event_id").Unique(),
	}
}

type AlertEpisode struct {
	ent.Schema
}

func (AlertEpisode) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
		KnowledgeEntityLinkMixin{},
	}
}

func (AlertEpisode) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("alert_definition_id", uuid.UUID{}).Immutable(),
		field.UUID("situation_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("status").Values("open", "closed").Default("open"),
		field.Time("started_at"),
		field.Time("last_observed_at"),
		field.Time("closed_at").Optional().Nillable(),
	}
}

func (AlertEpisode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("alert_definition", AlertDefinition.Type).
			Ref("episodes").
			Required().
			Unique().
			Immutable().
			Field("alert_definition_id"),
		edge.To("instances", AlertInstance.Type),
		edge.From("situation", Situation.Type).
			Ref("alert_episodes").
			Unique().
			Field("situation_id"),
	}
}

func (AlertEpisode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "alert_definition_id").
			Unique().
			Annotations(entsql.IndexWhere("status = 'open'")),
		index.Fields("tenant_id", "situation_id"),
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
