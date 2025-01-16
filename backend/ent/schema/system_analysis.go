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
		field.UUID("incident_id", uuid.UUID{}).Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SystemAnalysis) Edges() []ent.Edge {
	return []ent.Edge{
		// Hierarchical relationships
		//edge.To("children", SystemComponent.Type).
		//	From("parent").
		//	Unique(),
		// Control relationships
		edge.From("components", SystemComponent.Type).
			Ref("analyses").
			Through("analysis_components", SystemAnalysisComponent.Type),
		//edge.To("controls", SystemComponent.Type).
		//	StorageKey(edge.Table("system_component_control_relationship"), edge.Columns("controller_id", "controlled_id")).
		//	Through("control_relationships", SystemComponentControlRelationship.Type),
		//// Feedback relationships
		//edge.To("feedback_to", SystemComponent.Type).
		//	StorageKey(edge.Table("system_component_feedback_relationship"), edge.Columns("source_id", "target_id")).
		//	Through("feedback_relationships", SystemComponentFeedbackRelationship.Type),
		// Incident involvement
		//edge.From("incidents", Incident.Type).
		//	Ref("system_components").
		//	Through("incident_system_components", IncidentSystemComponent.Type),
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
