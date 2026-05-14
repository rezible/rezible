package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type ProviderEventSyncCursor struct {
	ent.Schema
}

func (ProviderEventSyncCursor) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (ProviderEventSyncCursor) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("provider").NotEmpty(),
		field.String("provider_source").NotEmpty(),
		field.String("cursor").Optional(),
		field.Time("last_synced_at").Default(time.Now),
	}
}

func (ProviderEventSyncCursor) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "provider", "provider_source").Unique(),
	}
}

type ProviderEventSyncRun struct {
	ent.Schema
}

func (ProviderEventSyncRun) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (ProviderEventSyncRun) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("provider").NotEmpty(),
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

func (ProviderEventSyncRun) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "provider", "started_at"),
		index.Fields("tenant_id", "status", "started_at"),
	}
}
