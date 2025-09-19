package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Event struct {
	ent.Schema
}

func (Event) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

var eventKind = []string{"alert", "interrupt", "message", "other"}

func (Event) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("provider_id"),
		field.Time("timestamp"),
		field.Enum("kind").Values(eventKind...),
		field.String("title"),
		field.String("description"),
		field.String("source"),
	}
}

func (Event) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("alert_instance", AlertInstance.Type).Ref("event"),
		edge.From("incident_event", IncidentEvent.Type).Ref("event"),
		edge.From("annotations", EventAnnotation.Type).Ref("event"),
	}
}
