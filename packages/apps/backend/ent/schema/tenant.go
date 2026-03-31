package schema

import (
	"entgo.io/ent"
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
	return []ent.Field{}
}

func (Tenant) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			rules.AllowIfSystemRole(),
			privacy.AlwaysDenyRule(),
		},
	}
}
