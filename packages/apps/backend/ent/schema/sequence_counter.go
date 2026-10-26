package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// SequenceCounter holds the last number handed out in a tenant's sequence of one kind.
type SequenceCounter struct {
	ent.Schema
}

func (SequenceCounter) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SequenceCounter) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Enum("kind").Values("incident", "situation"),
		field.Int("value").Positive(),
	}
}

func (SequenceCounter) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "kind").
			Unique().
			StorageKey("sequence_counters_kind_key"),
	}
}
