package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type OncallAnnotation struct {
	ent.Schema
}

func (OncallAnnotation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (OncallAnnotation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("event_id", uuid.UUID{}),
		field.UUID("roster_id", uuid.UUID{}),
		field.UUID("creator_id", uuid.UUID{}),
		field.Time("created_at").Default(time.Now),
		field.Int("minutes_occupied"),
		field.Text("notes"),
		field.JSON("tags", []string{}),
	}
}

func (OncallAnnotation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", OncallEvent.Type).Unique().Required().Field("event_id"),
		edge.To("roster", OncallRoster.Type).Unique().Required().Field("roster_id"),
		edge.To("creator", User.Type).Unique().Required().Field("creator_id"),

		edge.To("alert_feedback", AlertFeedback.Type).Unique(),
		edge.From("handovers", OncallShiftHandover.Type).Ref("pinned_annotations"),
	}
}
