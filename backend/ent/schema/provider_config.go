package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// ProviderConfig holds the schema definition for the ProviderConfig entity.
type ProviderConfig struct {
	ent.Schema
}

var ProviderTypes = []string{"auth_session", "users", "teams", "chat", "incidents", "oncall", "alerts", "ai", "system_components", "tickets", "playbooks"}

// Fields of the ProviderConfig.
func (ProviderConfig) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.Enum("provider_type").Values(ProviderTypes...),
		field.String("provider_name"),
		field.Bytes("provider_config"),
		field.Bool("enabled").Default(true),
		field.Time("updated_at").Default(time.Now),
	}
}

func (ProviderConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider_name", "provider_type").Unique(),
	}
}
