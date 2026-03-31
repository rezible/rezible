package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Integration holds the schema definition for the Integration entity.
type Integration struct {
	ent.Schema
}

func (Integration) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AccessScopeMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

// Fields of the Integration.
func (Integration) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("name"),
		field.JSON("config", map[string]any{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("user_preferences", map[string]any{}).
			Optional().Default(map[string]any{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
	}
}

func (Integration) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "name").Unique(),
	}
}

type IntegrationOAuthState struct {
	ent.Schema
}

func (IntegrationOAuthState) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AccessScopeMixin{},
		TenantMixin{},
	}
}

func (IntegrationOAuthState) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("user_id", uuid.New()).Default(uuid.New),
		field.String("state"),
		field.String("integration_name"),
		field.Time("expires_at").Default(func() time.Time {
			return time.Now().Add(time.Minute * 10)
		}),
	}
}

func (IntegrationOAuthState) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).Unique().Required().Field("user_id"),
	}
}
