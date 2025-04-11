package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

// OncallEventAnnotation holds the schema definition for the OncallEventAnnotation entity.
type OncallEventAnnotation struct {
	ent.Schema
}

// Fields of the OncallEventAnnotation.
func (OncallEventAnnotation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("event_id"),
		field.Time("created_at").Default(time.Now),
		field.Int("minutes_occupied"),
		field.Text("notes"),
		field.Bool("pinned"),
	}
}

// Edges of the OncallEventAnnotation.
func (OncallEventAnnotation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("shifts", OncallUserShift.Type).Ref("annotations"),
	}
}
