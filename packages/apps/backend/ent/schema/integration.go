package schema

import (
	"encoding/json"

	"entgo.io/ent"
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
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

// Fields of the Integration.
func (Integration) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("provider").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("display_name"),
		field.String("provider_installation_ref").NotEmpty(),
		field.JSON("installation_config", json.RawMessage{}).
			SchemaType(schemaTypeJsonB),
		field.JSON("user_settings", map[string]any{}).
			Optional().Default(map[string]any{}).
			SchemaType(schemaTypeJsonB),
	}
}

func (Integration) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "provider", "name", "provider_installation_ref").Unique(),
		index.Fields("tenant_id", "provider", "name"),
	}
}

type IntegrationUserInstallState struct {
	ent.Schema
}

func (IntegrationUserInstallState) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (IntegrationUserInstallState) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("user_id", uuid.New()).Default(uuid.New),
		field.String("integration_name"),
		field.Time("expires_at"),
		field.String("oauth_state").
			Optional(),
		field.JSON("installation_target_configs", map[string]json.RawMessage{}).
			Optional().
			SchemaType(schemaTypeJsonB),
	}
}

func (IntegrationUserInstallState) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "user_id", "integration_name").Unique(),
	}
}

func (IntegrationUserInstallState) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required().
			Field("user_id"),
	}
}
