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
		field.UUID("roster_id", uuid.UUID{}),
		field.UUID("creator_id", uuid.UUID{}),
		field.String("event_id"),
		field.Time("created_at").Default(time.Now),
		field.Int("minutes_occupied"),
		field.Text("notes"),
	}
}

// Edges of the OncallEventAnnotation.
func (OncallEventAnnotation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("roster", OncallRoster.Type).Unique().Required().Field("roster_id"),
		edge.To("creator", User.Type).Unique().Required().Field("creator_id"),

		edge.From("handovers", OncallUserShiftHandover.Type).Ref("pinned_annotations"),
	}
}
