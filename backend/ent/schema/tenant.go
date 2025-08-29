package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/privacy"
	"github.com/rezible/rezible/ent/schema/rules"
)

type Tenant struct {
	ent.Schema
}

func (Tenant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.UUID("public_id", uuid.UUID{}).Default(uuid.New),
		field.String("provider_id").NotEmpty(),
		field.Time("initial_setup_at").Optional(),
	}
}

func (Tenant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("public_id"),
	}
}

func (Tenant) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			rules.AllowIfSystemRole(),
			privacy.AlwaysDenyRule(),
		},
	}
}
