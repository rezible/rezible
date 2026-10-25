package schema

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent/schema/schematypes"
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
		field.Bytes("installation_config").
			GoType(json.RawMessage{}).
			ValueScanner(schematypes.EncryptedBytes[json.RawMessage]{}).
			Sensitive(),
		field.JSON("user_settings", json.RawMessage{}).
			Optional().
			SchemaType(schemaTypeJsonB),
		// WebhookTokenHash is the SHA-256 of the installation's webhook URL token. Only handlers that read it
		// give it meaning.
		field.Bytes("webhook_token_hash").
			Optional().
			Nillable().
			Sensitive(),
	}
}

func (Integration) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "provider", "name", "provider_installation_ref").Unique(),
		index.Fields("tenant_id", "provider", "name"),
		index.Fields("tenant_id", "webhook_token_hash").
			Unique().
			StorageKey("integrations_webhook_token_hash_key"),
		// A webhook delivery finds its integration by token alone.
		index.Fields("webhook_token_hash"),
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
		// InstallationTargetConfigs is the JSON-encoded map of target display names to installation configs.
		field.Bytes("installation_target_configs").
			Optional().
			ValueScanner(schematypes.EncryptedBytes[[]byte]{}).
			Sensitive(),
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
