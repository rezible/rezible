package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// IncidentType holds the schema definition for the IncidentType entity.
type IncidentType struct {
	ent.Schema
}

func (IncidentType) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
		ArchiveMixin{},
	}
}

// Fields of the IncidentType.
func (IncidentType) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
	}
}

// Edges of the IncidentType.
func (IncidentType) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incidents", Incident.Type).Ref("type"),
		edge.From("debrief_questions", IncidentDebriefQuestion.Type).Ref("incident_types"),
	}
}

func (IncidentType) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "name").Unique(),
	}
}
