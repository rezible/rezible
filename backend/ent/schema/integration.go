package schema

import (
	"time"

	"entgo.io/ent"
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
	}
}

var IntegrationTypes = []string{"chat", "users", "teams", "incidents", "oncall", "alerts", "system_components", "tickets", "playbooks"}

// Fields of the Integration.
func (Integration) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("name"),
		field.Enum("integration_type").Values(IntegrationTypes...),
		field.Bytes("config"),
		field.Bool("enabled").Default(true),
		field.Time("updated_at").Default(time.Now),
	}
}

func (Integration) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "name", "integration_type").Unique(),
	}
}
