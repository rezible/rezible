package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
		field.String("auth_provider_id"),
		field.String("name"),
		field.Time("initial_setup_at").Optional(),
	}
}

func (Organization) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "auth_provider_id").Unique(),
	}
}

func (Organization) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("preferences", OrganizationPreferences.Type).Ref("organization"),
		edge.From("roles", OrganizationRole.Type).Ref("organization"),
	}
}

type OrganizationPreferences struct {
	ent.Schema
}

func (OrganizationPreferences) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (OrganizationPreferences) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("org_id", uuid.UUID{}),
	}
}

func (OrganizationPreferences) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("organization", Organization.Type).
			Required().
			Unique().
			Field("org_id"),
	}
}

func (OrganizationPreferences) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "org_id").Unique(),
	}
}

type OrganizationRole struct {
	ent.Schema
}

func (OrganizationRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (OrganizationRole) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("org_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("role").Values("admin", "member").Default("member"),
	}
}

func (OrganizationRole) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("organization", Organization.Type).
			Required().
			Unique().
			Field("org_id"),
		edge.From("user", User.Type).
			Ref("organization_role").
			Required().
			Unique().
			Field("user_id"),
	}
}

func (OrganizationRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("org_id", "user_id").Unique(),
	}
}
