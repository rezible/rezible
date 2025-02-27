package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

type SystemComponentRelationship struct {
	ent.Schema
}

func (SystemComponentRelationship) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("provider_id").Optional(),
		field.UUID("source_id", uuid.UUID{}),
		field.UUID("target_id", uuid.UUID{}),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemComponentRelationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("source", SystemComponent.Type).
			Required().Unique().
			Field("source_id"),
		edge.To("target", SystemComponent.Type).
			Required().Unique().
			Field("target_id"),

		edge.From("system_analyses", SystemAnalysisRelationship.Type).
			Ref("component_relationship"),
	}
}
