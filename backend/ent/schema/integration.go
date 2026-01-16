package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
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
		BaseMixin{},
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
			SchemaType(map[string]string{
				dialect.Postgres: "jsonb",
			}),
		field.JSON("user_config", map[string]any{}).Optional(),
	}
}

func (Integration) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "name").Unique(),
	}
}
