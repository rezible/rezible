package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

// OncallHandoverTemplate holds the schema definition for the OncallHandoverTemplate entity.
type OncallHandoverTemplate struct {
	ent.Schema
}

// Fields of the OncallHandoverTemplate.
func (OncallHandoverTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now),
		field.Bytes("contents"),
		field.Bool("is_default").Default(false),
	}
}

// Edges of the OncallHandoverTemplate.
func (OncallHandoverTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("roster", OncallRoster.Type),
	}
}
