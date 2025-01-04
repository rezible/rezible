package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

type IncidentEventContributingFactor struct {
	ent.Schema
}

func (IncidentEventContributingFactor) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.String("factor_type").
			NotEmpty(),
		field.Text("description").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (IncidentEventContributingFactor) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", IncidentEvent.Type).
			Ref("factors").
			Unique().
			Required(),
	}
}
