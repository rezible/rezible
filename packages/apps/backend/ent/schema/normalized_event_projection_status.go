package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

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
