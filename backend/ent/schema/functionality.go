package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Functionality holds the schema definition for the Functionality entity.
type Functionality struct {
	ent.Schema
}

// Fields of the Functionality.
func (Functionality) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("name"),
	}
}

func (Functionality) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Edges of the Functionality.
func (Functionality) Edges() []ent.Edge {
	return []ent.Edge{
		//edge.To("incidents", IncidentResourceImpact.Type),
	}
}
