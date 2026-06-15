package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type IncidentImpact struct {
	ent.Schema
}

func (IncidentImpact) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (IncidentImpact) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.UUID("knowledge_entity_id", uuid.UUID{}),
		field.String("source").Optional(),
		field.Text("note").Optional(),
	}
}

func (IncidentImpact) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("incident", Incident.Type).
			Required().
			Unique().
			Field("incident_id"),
		edge.To("knowledge_entity", KnowledgeEntity.Type).
			Required().
			Unique().
			Field("knowledge_entity_id"),
	}
}

func (IncidentImpact) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "incident_id"),
		index.Fields("tenant_id", "knowledge_entity_id"),
		index.Fields("tenant_id", "incident_id", "knowledge_entity_id").Unique(),
	}
}
