package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/schema/schematypes"
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
		field.Int("resolution_timeout_seconds").
			NonNegative().
			Comment("How long after its last firing notification a window is assumed ended; 0 never"),
		field.JSON("identity_group_labels", []string{}).
			Optional().
			Comment("Label names that group instances"),
		field.Time("metadata_observed_at").
			Comment("occurred_at of the notification whose metadata is stored"),
		field.String("metadata_event_ref").
			Comment("provider_event_ref of that notification, which breaks ties"),
		field.UUID("situation_signal_attention_id", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

func (AlertDefinition) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("playbooks", Playbook.Type).Ref("alert_definitions"),
		edge.To("episodes", AlertEpisode.Type),
		edge.To("situation_signal_attention", SituationSignalAttention.Type).
			Unique().
			Field("situation_signal_attention_id"),
	}
}

func (AlertDefinition) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("situation_signal_attention_id").Unique(),
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
		field.String("instance_key").
			Immutable().
			Comment("The source instance's identity"),
		field.String("grouping_key").
			Immutable().
			Comment("Grouping for detection counts"),
		field.JSON("labels", map[string]string{}).
			Optional().
			Comment("From the window's newest notification"),
		field.String("summary").
			Optional().
			Comment("From the window's newest notification"),
		field.Enum("severity").
			GoType(schematypes.SignalSeverityUnknown).
			Default(string(schematypes.SignalSeverityUnknown)).
			Comment("Peak severity of the window's notifications"),
		field.Time("fired_at").
			Immutable().
			Comment("The provider's window start"),
		field.Time("last_observed_at").
			Comment("Latest firing notification; a timeout counts from it"),
		field.Time("resolved_at").
			Optional().
			Nillable().
			Comment("Latest reported resolution"),
		field.Time("ended_at").
			Optional().
			Nillable().
			Comment("Effective end; nil while firing"),
		field.Enum("end_reason").
			Values("resolved", "superseded", "timeout").
			Optional().
			Nillable(),
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
		edge.To("events", AlertInstanceEvent.Type),
		edge.From("feedback", AlertFeedback.Type).
			Ref("alert_instance"),
	}
}

func (AlertInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "alert_episode_id", "instance_key", "fired_at").Unique(),
	}
}

// AlertInstanceEvent links a notification's normalized event to its window.
type AlertInstanceEvent struct {
	ent.Schema
}

func (AlertInstanceEvent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (AlertInstanceEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("alert_instance_id", uuid.UUID{}).Immutable(),
		field.UUID("event_id", uuid.UUID{}).Immutable(),
	}
}

func (AlertInstanceEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("instance", AlertInstance.Type).
			Ref("events").
			Required().
			Unique().
			Immutable().
			Field("alert_instance_id"),
		edge.To("event", NormalizedEvent.Type).
			Required().
			Unique().
			Immutable().
			Field("event_id"),
	}
}

func (AlertInstanceEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "event_id").Unique(),
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
		field.Time("started_at").
			Comment("Earliest window start"),
		field.Time("closed_at").
			Optional().
			Nillable().
			Comment("Effective closure; nil while open"),
		field.Enum("highest_severity").
			GoType(schematypes.SignalSeverityUnknown).
			Default(string(schematypes.SignalSeverityUnknown)).
			Comment("Highest severity of the episode's windows"),
		field.JSON("identity_group_labels", []string{}).
			Optional().
			Immutable().
			Comment("The definition's identity group labels when the episode opened"),
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
	}
}

func (AlertEpisode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "alert_definition_id", "started_at"),
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
