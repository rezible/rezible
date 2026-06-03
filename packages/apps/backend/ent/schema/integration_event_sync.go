package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type IntegrationEventSyncCursor struct {
	ent.Schema
}

func (IntegrationEventSyncCursor) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (IntegrationEventSyncCursor) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("integration_id", uuid.UUID{}),
		field.String("provider_source").NotEmpty(),
		field.String("cursor").Optional(),
		field.Time("last_synced_at").Default(time.Now),
	}
}

func (IntegrationEventSyncCursor) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "integration_id", "provider_source").Unique(),
	}
}

func (IntegrationEventSyncCursor) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("integration", Integration.Type).
			Unique().
			Required().
			Field("integration_id"),
	}
}

type IntegrationEventSyncRun struct {
	ent.Schema
}

func (IntegrationEventSyncRun) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (IntegrationEventSyncRun) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("integration_id", uuid.UUID{}),
		field.JSON("source_cursors", map[string]string{}).Optional(),
		field.String("sync_reason").Default("manual"),
		field.Time("started_at").Default(time.Now),
		field.Time("finished_at").Optional().Nillable(),
		field.Enum("status").Values("success", "failed", "skipped"),
		field.Int("events_pulled").Default(0),
		field.Int("events_ingested").Default(0),
		field.Int("duplicates").Default(0),
		field.String("failure_message").Optional(),
	}
}

func (IntegrationEventSyncRun) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "integration_id", "started_at"),
		index.Fields("tenant_id", "status", "started_at"),
	}
}

func (IntegrationEventSyncRun) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("integration", Integration.Type).
			Unique().
			Required().
			Field("integration_id"),
	}
}
