package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
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
		edge.From("analysis_entities", SystemAnalysisEntity.Type).
			Ref("analysis"),
		edge.From("analysis_relationships", SystemAnalysisRelationship.Type).
			Ref("analysis"),
		edge.From("entries", SystemAnalysisEntry.Type).
			Ref("analysis"),

		edge.From("agent_sessions", AgentSession.Type).Ref("system_analysis"),
		edge.To("situation_investigation", SituationInvestigation.Type).Unique(),
	}
}

func (SystemAnalysis) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "scope_entity_id"),
		index.Fields("tenant_id", "subject_entity_id"),
	}
}

type SystemAnalysisEntity struct {
	ent.Schema
}

func (SystemAnalysisEntity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SystemAnalysisEntity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("analysis_id", uuid.UUID{}).
			Immutable(),
		field.UUID("knowledge_entity_id", uuid.UUID{}).
			Immutable(),
		field.Float("pos_x").
			Optional().
			Nillable(),
		field.Float("pos_y").
			Optional().
			Nillable(),
		field.Bool("hidden").
			Default(false),
		field.String("label_override").
			Optional().
			Nillable(),
		field.Text("description_override").
			Optional().
			Nillable(),
		field.JSON("properties", map[string]any{}).
			Optional().
			SchemaType(schemaTypeJsonB),
	}
}

func (SystemAnalysisEntity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("analysis", SystemAnalysis.Type).
			Required().
			Unique().
			Immutable().
			Field("analysis_id"),
		edge.To("knowledge_entity", KnowledgeEntity.Type).
			Required().
			Unique().
			Immutable().
			Field("knowledge_entity_id"),
	}
}

func (SystemAnalysisEntity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "analysis_id", "knowledge_entity_id").Unique(),
		index.Fields("tenant_id", "knowledge_entity_id"),
	}
}

type SystemAnalysisRelationship struct {
	ent.Schema
}

func (SystemAnalysisRelationship) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SystemAnalysisRelationship) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("analysis_id", uuid.UUID{}).
			Immutable(),
		field.UUID("knowledge_relationship_id", uuid.UUID{}).
			Immutable(),
		field.Bool("hidden").
			Default(false),
		field.String("label_override").
			Optional().
			Nillable(),
		field.Text("description_override").
			Optional().
			Nillable(),
		field.JSON("layout", map[string]any{}).
			Optional().
			SchemaType(schemaTypeJsonB),
		field.JSON("properties", map[string]any{}).
			Optional().
			SchemaType(schemaTypeJsonB),
	}
}

func (SystemAnalysisRelationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("analysis", SystemAnalysis.Type).
			Required().
			Unique().
			Immutable().
			Field("analysis_id"),
		edge.To("knowledge_relationship", KnowledgeRelationship.Type).
			Required().
			Unique().
			Immutable().
			Field("knowledge_relationship_id"),
	}
}

func (SystemAnalysisRelationship) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "analysis_id", "knowledge_relationship_id").Unique(),
		index.Fields("tenant_id", "knowledge_relationship_id"),
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
		field.UUID("analysis_id", uuid.UUID{}).
			Immutable(),
		field.String("reference").
			Optional().
			Nillable().
			Immutable(),
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
			Immutable().
			Field("analysis_id"),
		edge.From("subjects", SystemAnalysisEntrySubject.Type).
			Ref("entry"),
	}
}

func (SystemAnalysisEntry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "analysis_id", "reference").Unique(),
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
		field.UUID("entry_id", uuid.UUID{}).
			Immutable(),
		field.UUID("knowledge_entity_id", uuid.UUID{}).
			Optional().
			Nillable().
			Immutable(),
		field.UUID("knowledge_relationship_id", uuid.UUID{}).
			Optional().
			Nillable().
			Immutable(),
		field.UUID("knowledge_evidence_id", uuid.UUID{}).
			Optional().
			Nillable().
			Immutable(),
		field.String("role").NotEmpty().
			Comment("How the graph subject participates in the analysis entry, e.g. primary, affected, contributing, evidence_for."),
	}
}

func (SystemAnalysisEntrySubject) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Checks: map[string]string{
			"system_analysis_entry_subject_exactly_one_reference": "num_nonnulls(knowledge_entity_id, knowledge_relationship_id, knowledge_evidence_id) = 1",
		}},
	}
}

func (SystemAnalysisEntrySubject) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entry", SystemAnalysisEntry.Type).
			Required().
			Unique().
			Immutable().
			Field("entry_id"),
		edge.To("knowledge_entity", KnowledgeEntity.Type).
			Unique().
			Immutable().
			Field("knowledge_entity_id"),
		edge.To("knowledge_relationship", KnowledgeRelationship.Type).
			Unique().
			Immutable().
			Field("knowledge_relationship_id"),
		edge.To("knowledge_evidence", KnowledgeEvidence.Type).
			Unique().
			Immutable().
			Field("knowledge_evidence_id"),
	}
}

func (SystemAnalysisEntrySubject) Indexes() []ent.Index {
	return []ent.Index{
		// TODO: clean these up
		index.Fields("tenant_id", "entry_id", "knowledge_entity_id", "role").Unique(),
		index.Fields("tenant_id", "entry_id", "knowledge_relationship_id", "role").Unique(),
		index.Fields("tenant_id", "entry_id", "knowledge_evidence_id", "role").Unique(),
		index.Fields("tenant_id", "knowledge_entity_id"),
		index.Fields("tenant_id", "knowledge_relationship_id"),
		index.Fields("tenant_id", "knowledge_evidence_id"),
	}
}
