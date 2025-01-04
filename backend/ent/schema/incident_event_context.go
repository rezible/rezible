package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

type IncidentEventContext struct {
	ent.Schema
}

func (IncidentEventContext) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.Text("system_state").
			Optional(),
		field.JSON("decision_options", []string{}).
			Optional(),
		field.Text("decision_rationale").
			Optional(),
		field.JSON("involved_personnel", []string{}).
			Optional(),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (IncidentEventContext) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", IncidentEvent.Type).
			Ref("context").
			Unique().
			Required(),
	}
}
