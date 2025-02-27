package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

type SystemAnalysis struct {
	ent.Schema
}

func (SystemAnalysis) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SystemAnalysis) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("incident", Incident.Type).
			Unique().Required().Field("incident_id"),
		edge.From("components", SystemComponent.Type).
			Ref("system_analyses").
			Through("analysis_components", SystemAnalysisComponent.Type),
		edge.From("relationships", SystemAnalysisRelationship.Type).
			Ref("system_analysis"),
	}
}

type SystemAnalysisComponent struct {
	ent.Schema
}

func (SystemAnalysisComponent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("analysis_id", uuid.UUID{}),
		field.UUID("component_id", uuid.UUID{}),
		field.Text("description").Optional(),
		field.Int("pos_x").Default(0),
		field.Int("pos_y").Default(0),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemAnalysisComponent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("analysis", SystemAnalysis.Type).
			Required().Unique().Field("analysis_id"),
		edge.To("component", SystemComponent.Type).
			Required().Unique().Field("component_id"),
	}
}
