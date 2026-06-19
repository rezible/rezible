package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type UserAuthSession struct {
	ent.Schema
}

func (UserAuthSession) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (UserAuthSession) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("user_id", uuid.New()),
		field.UUID("organization_id", uuid.New()),
		field.Time("expires_at"),
		field.Strings("scopes").Optional(),
	}
}

func (UserAuthSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).Unique().Required().Field("user_id"),
		edge.To("organization", Organization.Type).Unique().Required().Field("organization_id"),
	}
}

func (UserAuthSession) Indexes() []ent.Index {
	return []ent.Index{}
}
