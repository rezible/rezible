package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type SystemHazard struct {
	ent.Schema
}

func (SystemHazard) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SystemHazard) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").NotEmpty(),
		field.Text("description"),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SystemHazard) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("components", SystemComponent.Type),
		edge.To("constraints", SystemComponentConstraint.Type),
		edge.To("relationships", SystemComponentRelationship.Type),
	}
}
