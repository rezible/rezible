package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

type OncallAnnotation struct {
	ent.Schema
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
	}
}

func (OncallAnnotation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", OncallEvent.Type).Unique().Required().Field("event_id"),
		edge.To("roster", OncallRoster.Type).Unique().Required().Field("roster_id"),
		edge.To("creator", User.Type).Unique().Required().Field("creator_id"),

		edge.To("alert_feedback", OncallAnnotationAlertFeedback.Type).Unique(),
		edge.From("handovers", OncallUserShiftHandover.Type).Ref("pinned_annotations"),
	}
}

// OncallAnnotationAlertFeedback holds the schema definition for the OncallAnnotationAlertFeedback entity.
type OncallAnnotationAlertFeedback struct {
	ent.Schema
}

// Fields of the OncallAnnotationAlertFeedback.
func (OncallAnnotationAlertFeedback) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("annotation_id", uuid.UUID{}),
		field.Bool("actionable"),
		field.Bool("documentation_available"),
		field.Enum("accuracy").Values("yes", "no", "unknown"),
	}
}

// Edges of the OncallAnnotationAlertFeedback.
func (OncallAnnotationAlertFeedback) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("annotation", OncallAnnotation.Type).
			Ref("alert_feedback").
			Field("annotation_id").Unique().Required(),
	}
}
