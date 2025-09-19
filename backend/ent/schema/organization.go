package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Organization struct {
	ent.Schema
}

func (Organization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (Organization) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("provider_id").NotEmpty(),
		field.String("name"),
		field.Time("initial_setup_at").Optional(),
	}
}

func (Organization) Edges() []ent.Edge {
	return []ent.Edge{}
}
