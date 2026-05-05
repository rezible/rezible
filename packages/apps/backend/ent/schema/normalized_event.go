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
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("provider").NotEmpty(),
		field.String("source").NotEmpty(),
		field.Enum("kind").Values(
			"chat_message",
			"repository_observed",
			"change_event_observed",
		),
		field.String("subject_kind").NotEmpty(),
		field.String("subject_external_ref").NotEmpty(),
		field.String("source_event_key").NotEmpty(),
		field.String("dedupe_key").Optional(),
		field.Time("occurred_at"),
		field.Time("received_at"),
		field.String("processing_version").NotEmpty(),
		field.JSON("attributes", map[string]any{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("created_at").Default(time.Now),
	}
}

func (NormalizedEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "provider", "source", "processing_version", "source_event_key", "kind", "subject_external_ref").
			Unique(),
		index.Fields("tenant_id", "kind", "occurred_at"),
	}
}
