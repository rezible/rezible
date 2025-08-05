package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

// ProviderSyncHistory holds the schema definition for the ProviderSyncHistory entity.
type ProviderSyncHistory struct {
	ent.Schema
}

func (ProviderSyncHistory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

// Fields of the ProviderSyncHistory.
func (ProviderSyncHistory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("data_type"),
		field.Time("started_at").Default(time.Now),
		field.Time("finished_at").Default(time.Now),
		field.Int("num_mutations"),
	}
}

//func (ProviderSyncHistory) Indexes() []ent.Index {
//	return []ent.Index{
//		index.Fields("provider_name", "provider_type").Unique(),
//	}
//}
