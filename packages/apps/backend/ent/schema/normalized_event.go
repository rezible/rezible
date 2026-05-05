package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
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
		field.UUID("id", uuid.UUID{}).Default(uuid.New).
			Comment("Internal identifier for this normalized event record."),
		field.String("provider").NotEmpty().
			Comment("Integration provider that produced the event, such as slack or github."),
		field.String("provider_source").NotEmpty().
			Comment("Provider-specific event stream or webhook source the event came from."),
		field.Enum("kind").Values(
			"chat_message",
			"repository_observed",
			"change_event_observed",
		).Comment("Normalized event type used to select validation and projection behavior."),
		field.String("subject_kind").NotEmpty().
			Comment("Provider-neutral type of the primary subject this event is about."),
		field.String("subject_ref").NotEmpty().
			Comment("Stable external reference for the primary subject this event is about."),
		field.String("provider_event_ref").NotEmpty().
			Comment("Stable provider reference for the source event, used with the provider fields for idempotency."),
		field.String("dedupe_key").Optional().
			Comment("Optional ingestion dedupe key from the upstream provider event pipeline."),
		field.Time("occurred_at").
			Comment("Time the event occurred according to the provider or normalized payload."),
		field.Time("received_at").
			Comment("Time the raw provider event was received by Rezible."),
		field.String("processing_version").NotEmpty().
			Comment("Normalizer version used to produce this event shape and attributes."),
		field.JSON("attributes", map[string]any{}).
			Comment("Validated normalized attributes for this event kind.").
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("created_at").Default(time.Now).
			Comment("Time this normalized event record was persisted."),
	}
}

func (NormalizedEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "provider", "provider_source", "processing_version", "provider_event_ref", "kind", "subject_ref").
			Unique(),
		index.Fields("tenant_id", "kind", "occurred_at"),
	}
}
