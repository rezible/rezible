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

// Fields of the Integration.
func (Integration) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("name"),
		field.Bytes("config"),
		field.Time("updated_at").Default(time.Now),
	}
}

func (Integration) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "name").Unique(),
	}
}
