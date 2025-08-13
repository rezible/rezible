package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Alert holds the schema definition for the Alert entity.
type Alert struct {
	ent.Schema
}

// Fields of the Alert.
func (Alert) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("provider_id"),
		field.String("title"),
		field.String("description").Optional(),
		field.String("definition").Optional(),
		field.UUID("roster_id", uuid.UUID{}).Optional(),
	}
}

func (Alert) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

// Edges of the Alert.
func (Alert) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("playbooks", Playbook.Type).Ref("alerts"),
		edge.From("roster", OncallRoster.Type).Ref("alerts").Unique().Field("roster_id"),
		edge.From("events", OncallEvent.Type).Ref("alert"),
		edge.To("feedback", AlertFeedback.Type),
	}
}

// AlertFeedback holds the schema definition for the AlertFeedback entity.
type AlertFeedback struct {
	ent.Schema
}

func (AlertFeedback) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

// Fields of the AlertFeedback.
func (AlertFeedback) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("alert_id", uuid.UUID{}),
		field.UUID("annotation_id", uuid.UUID{}),
		field.Bool("actionable"),
		field.Enum("accurate").Values("yes", "no", "unknown"),
		field.Bool("documentation_available"),
		field.Bool("documentation_needs_update"),
	}
}

// Edges of the OncallAnnotationAlertFeedback.
func (AlertFeedback) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("alert", Alert.Type).
			Ref("feedback").
			Required().
			Unique().
			Field("alert_id"),
		edge.From("annotation", OncallAnnotation.Type).
			Ref("alert_feedback").
			Required().
			Unique().
			Field("annotation_id"),
	}
}

type AlertMetrics struct {
	ent.View
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

/*
type AlertMetricsSummary struct {
	ent.View
}

const AlertMetricsSummarySchema = `CREATE VIEW
"alert_metrics_summary" ("alert_id","event_id","event_roster_id","annotation_roster_id","tenant_id","event_timestamp","fb_accurate","fb_actionable","fb_documentation")
AS
SELECT
    a.id AS alert_id,
    e.id AS event_id,
    e.roster_id AS event_roster_id,
    ann.roster_id AS annotation_roster_id,
    a.tenant_id AS tenant_id,
    e.timestamp AS event_timestamp,
    fb.accurate AS fb_accurate,
    fb.actionable AS fb_actionable,
    fb.documentation_available AS fb_documentation
FROM alerts a
    JOIN oncall_events e
        ON a.id=e.alert_id
    JOIN oncall_annotations ann
        ON ann.event_id=e.id
    JOIN alert_feedbacks fb
        ON fb.annotation_id=ann.id`

func (AlertMetricsSummary) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.ViewFor(dialect.Postgres, func(s *sql.Selector) {
			s.
				AppendSelectAs("e.alert_id", "alert_id").
				AppendSelectAs("e.id", "event_id").
				AppendSelectAs("e.roster_id", "event_roster_id").
				AppendSelectAs("ann.roster_id", "annotation_roster_id").
				AppendSelectAs("e.tenant_id", "tenant_id").
				AppendSelectAs("e.timestamp", "event_timestamp").
				AppendSelectAs("fb.accurate", "fb_accurate").
				AppendSelectAs("fb.actionable", "fb_actionable").
				AppendSelectAs("fb.documentation_available", "fb_documentation").
				From(sql.Table("oncall_events").As("e")).
				Join(sql.Table("oncall_annotations").As("ann")).On("e.id", "ann.event_id").
				Join(sql.Table("alert_feedbacks").As("fb")).On("ann.id", "fb.annotation_id")
		}),
	}
}

func (AlertMetricsSummary) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("event_id", uuid.UUID{}),
		field.UUID("event_roster_id", uuid.UUID{}),
		field.UUID("annotation_roster_id", uuid.UUID{}),
		field.Int("tenant_id"),
		field.Time("event_timestamp"),
		field.Bool("fb_accurate"),
		field.String("fb_actionable"),
		field.String("fb_documentation"),
	}
}

func (AlertMetricsSummary) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (AlertMetricsSummary) Policy() ent.Policy {
	return rules.FilterTenantRule()
}
*/
