package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

type SystemComponent struct {
	ent.Schema
}

func (SystemComponent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").NotEmpty(),
		field.String("provider_id").Optional(),
		field.UUID("kind_id", uuid.UUID{}).Optional(),
		field.Text("description").Optional(),
		field.JSON("properties", map[string]any{}).Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SystemComponent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("kind", SystemComponentKind.Type).
			Unique().
			Field("kind_id"),

		edge.To("related", SystemComponent.Type).
			StorageKey(
				edge.Table("system_component_relationships"),
				edge.Columns("source_id", "target_id")).
			Through("component_relationships", SystemComponentRelationship.Type),

		edge.To("system_analyses", SystemAnalysis.Type).
			Through("system_analysis_components", SystemAnalysisComponent.Type),
		//edge.To("system_analysis_relations", SystemComponent.Type).
		//	//StorageKey(
		//	//	edge.Table("system_analysis_relations"),
		//	//	edge.Columns("source_component_id", "target_component_id")).
		//	Through("system_analysis_relations", SystemAnalysisRelationship.Type),

		edge.From("events", IncidentEvent.Type).
			Ref("system_components").
			Through("event_components", IncidentEventSystemComponent.Type),
		edge.From("constraints", SystemComponentConstraint.Type).
			Ref("component"),
		edge.From("controls", SystemComponentControl.Type).
			Ref("component"),
		edge.From("signals", SystemComponentSignal.Type).
			Ref("component"),

		edge.From("hazards", SystemHazard.Type).Ref("components"),
	}
}

type SystemComponentKind struct {
	ent.Schema
}

func (SystemComponentKind) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("provider_id").Optional(),
		field.Text("label"),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemComponentKind) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("components", SystemComponent.Type).Ref("kind"),
	}
}

type SystemComponentConstraint struct {
	ent.Schema
}

func (SystemComponentConstraint) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("component_id", uuid.UUID{}),
		field.Text("label"),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemComponentConstraint) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("component", SystemComponent.Type).
			Required().Unique().Field("component_id"),

		edge.From("hazards", SystemHazard.Type).Ref("constraints"),
	}
}

type SystemComponentSignal struct {
	ent.Schema
}

func (SystemComponentSignal) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("component_id", uuid.UUID{}),
		field.Text("label"),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemComponentSignal) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("component", SystemComponent.Type).
			Required().Unique().Field("component_id"),
		edge.From("relationships", SystemAnalysisRelationship.Type).
			Ref("signals").
			Through("feedback_signals", SystemRelationshipFeedbackSignal.Type),
	}
}

type SystemComponentControl struct {
	ent.Schema
}

func (SystemComponentControl) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("component_id", uuid.UUID{}),
		field.Text("label"),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SystemComponentControl) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("component", SystemComponent.Type).
			Required().Unique().Field("component_id"),
		edge.From("relationships", SystemAnalysisRelationship.Type).
			Ref("controls").
			Through("control_actions", SystemRelationshipControlAction.Type),
	}
}
