package schema

import (
	"entgo.io/ent/schema/edge"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Playbook holds the schema definition for the Playbook entity.
type Playbook struct {
	ent.Schema
}

func (Playbook) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
	}
}

// Fields of the Playbook.
func (Playbook) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.Bytes("content"),
	}
}

// Edges of the Playbook.
func (Playbook) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("alert_definitions", AlertDefinition.Type),
	}
}
