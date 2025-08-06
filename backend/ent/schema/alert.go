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
		TenantMixin{},
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
		field.UUID("annotation_id", uuid.UUID{}),
		field.Bool("actionable"),
		field.Enum("accurate").Values("yes", "no", "unknown"),
		field.Enum("documentation_available").Values("yes", "needs_update", "no"),
	}
}

// Edges of the OncallAnnotationAlertFeedback.
func (AlertFeedback) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("annotation", OncallAnnotation.Type).
			Ref("alert_feedback").
			Field("annotation_id").
			Unique().
			Required(),
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
		TenantMixin{},
	}
}

// Edges of the AlertMetrics.
func (AlertMetrics) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("alert", Alert.Type).Unique().Required().Field("alert_id"),
	}
}
