package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/privacy"
	"github.com/rezible/rezible/ent/schema/privacyrules"
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
	}
}

func (Tenant) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			privacyrules.AllowIfSystem(),
			privacy.AlwaysDenyRule(),
		},
	}
}
