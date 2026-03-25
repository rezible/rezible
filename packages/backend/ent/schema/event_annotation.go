package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type EventAnnotation struct {
	ent.Schema
}

func (EventAnnotation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (EventAnnotation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("event_id", uuid.UUID{}),
		field.UUID("creator_id", uuid.UUID{}),
		field.Time("created_at").Default(time.Now),
		field.Int("minutes_occupied"),
		field.Text("notes"),
		field.JSON("tags", []string{}),
	}
}

func (EventAnnotation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", Event.Type).Unique().Required().Field("event_id"),
		edge.To("creator", User.Type).Unique().Required().Field("creator_id"),

		edge.From("handovers", OncallShiftHandover.Type).Ref("pinned_annotations"),
	}
}
