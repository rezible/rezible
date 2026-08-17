package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type SystemAnalysis struct {
	ent.Schema
}

func (SystemAnalysis) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SystemAnalysis) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("scope_entity_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Optional knowledge graph entity that bounds analysis context and default graph traversal."),
		field.UUID("subject_entity_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Optional primary knowledge graph entity this analysis is about."),
		field.Time("reference_time").
			Optional().
			Nillable().
			Comment("Optional evidence time used to render historical graph state; nil means current state."),
	}
}

func (SystemAnalysis) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("scope_entity", KnowledgeEntity.Type).
			Unique().
			Field("scope_entity_id"),
		edge.To("subject_entity", KnowledgeEntity.Type).
			Unique().
			Field("subject_entity_id"),
		edge.From("entries", SystemAnalysisEntry.Type).
			Ref("analysis"),
	}
}

func (SystemAnalysis) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "scope_entity_id"),
		index.Fields("tenant_id", "subject_entity_id"),
	}
}

type SystemAnalysisEntry struct {
	ent.Schema
}

func (SystemAnalysisEntry) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SystemAnalysisEntry) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("analysis_id", uuid.UUID{}),
		field.Enum("kind").
			Values("observation", "context", "decision", "action", "finding", "recommendation"),
		field.Time("occurred_at").
			Optional().
			Nillable().
			Comment("Domain time for observations/actions/events; nil for timeless findings or context."),
		field.Int("sequence").Default(0),
		field.String("title").NotEmpty(),
		field.Text("body").Optional(),
		field.JSON("properties", map[string]any{}).
			Optional().
			SchemaType(schemaTypeJsonB).
			Comment("Structured workflow-specific details that should not become core graph schema."),
	}
}

func (SystemAnalysisEntry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("analysis", SystemAnalysis.Type).
			Required().
			Unique().
			Field("analysis_id"),
		edge.From("subjects", SystemAnalysisEntrySubject.Type).
			Ref("entry"),
	}
}

func (SystemAnalysisEntry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "analysis_id", "kind"),
		index.Fields("tenant_id", "analysis_id", "sequence"),
	}
}

type SystemAnalysisEntrySubject struct {
	ent.Schema
}

func (SystemAnalysisEntrySubject) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SystemAnalysisEntrySubject) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("entry_id", uuid.UUID{}),
		field.UUID("knowledge_entity_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("knowledge_relationship_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("knowledge_evidence_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("role").NotEmpty().
			Comment("How the graph subject participates in the analysis entry, e.g. primary, affected, contributing, evidence_for."),
	}
}

func (SystemAnalysisEntrySubject) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entry", SystemAnalysisEntry.Type).
			Required().
			Unique().
			Field("entry_id"),
		edge.To("knowledge_entity", KnowledgeEntity.Type).
			Unique().
			Field("knowledge_entity_id"),
		edge.To("knowledge_relationship", KnowledgeRelationship.Type).
			Unique().
			Field("knowledge_relationship_id"),
		edge.To("knowledge_evidence", KnowledgeEvidence.Type).
			Unique().
			Field("knowledge_evidence_id"),
	}
}

func (SystemAnalysisEntrySubject) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "entry_id", "knowledge_entity_id", "role").Unique(),
		index.Fields("tenant_id", "entry_id", "knowledge_relationship_id", "role").Unique(),
		index.Fields("tenant_id", "entry_id", "knowledge_evidence_id", "role").Unique(),
		index.Fields("tenant_id", "knowledge_entity_id"),
		index.Fields("tenant_id", "knowledge_relationship_id"),
		index.Fields("tenant_id", "knowledge_evidence_id"),
	}
}
