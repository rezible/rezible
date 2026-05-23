package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type NormalizedEvent struct {
	ent.Schema
}

func (NormalizedEvent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (NormalizedEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Enum("activity_kind").Values("received", "observed", "deleted").
			Comment("Kind of activity represented by the event."),
		field.String("provider").NotEmpty().
			Comment("Integration provider that produced the event, such as slack or github."),
		field.String("provider_source").NotEmpty().
			Comment("Provider-specific event stream or webhook source the event came from."),
		field.String("provider_event_ref").NotEmpty().
			Comment("Stable provider reference for the source event, used with the provider fields for idempotency."),
		field.String("provider_subject_ref").NotEmpty().
			Comment("Stable provider reference for the primary subject this event is about."),
		field.String("subject_kind").
			Comment("Provider-neutral type of the primary subject this event is about."),
		field.JSON("attributes", map[string]any{}).
			Comment("Normalized attributes for this event kind.").
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("created_at").Default(time.Now),
		field.Time("occurred_at"),
		field.Time("received_at"),
	}
}

func (NormalizedEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("alert_feedback", AlertFeedback.Type),
	}
}

func (NormalizedEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "provider", "provider_source", "provider_event_ref", "provider_subject_ref").
			Unique(),
		index.Fields("tenant_id", "provider", "provider_source", "occurred_at"),
	}
}

type NormalizedEventProjectionStatus struct {
	ent.Schema
}

func (NormalizedEventProjectionStatus) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (NormalizedEventProjectionStatus) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("normalized_event_id", uuid.UUID{}),
		field.String("handler_name").NotEmpty(),
		field.Enum("status").Values("pending", "succeeded", "failed").Default("pending"),
		field.String("last_error").Optional(),
		field.Time("last_attempted_at").Optional().Nillable(),
		field.Time("succeeded_at").Optional().Nillable(),
		field.Time("failed_at").Optional().Nillable(),
	}
}

func (NormalizedEventProjectionStatus) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("normalized_event", NormalizedEvent.Type).
			Required().
			Unique().
			Field("normalized_event_id"),
	}
}

func (NormalizedEventProjectionStatus) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "normalized_event_id", "handler_name").Unique(),
		index.Fields("tenant_id", "status", "updated_at"),
	}
}

type EventAnnotation struct {
	ent.Schema
}

func (EventAnnotation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (EventAnnotation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("event_id", uuid.UUID{}),
		field.UUID("creator_id", uuid.UUID{}),
		field.Time("created_at").Default(time.Now),
		field.Int("minutes_occupied"),
		field.Text("notes"),
		field.JSON("tags", []string{}),
	}
}

func (EventAnnotation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", NormalizedEvent.Type).
			Unique().
			Required().
			Field("event_id"),
		edge.To("creator", User.Type).
			Unique().
			Required().
			Field("creator_id"),

		edge.From("handovers", OncallShiftHandover.Type).
			Ref("pinned_annotations"),
	}
}
