package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type AgentRun struct {
	ent.Schema
}

func (AgentRun) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentRun) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("owner_user_id", uuid.UUID{}),
		field.String("workflow").NotEmpty(),
		field.Time("started_at").Optional().Nillable(),
		field.Bytes("input"),
		field.JSON("metadata", map[string]any{}).
			SchemaType(schemaTypeJsonB).
			Optional(),
	}
}

func (AgentRun) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("owner_user", User.Type).
			Required().
			Unique().
			Field("owner_user_id"),
		edge.To("subjects", AgentRunSubject.Type),
		edge.To("result", AgentRunResult.Type).
			Unique(),
		edge.From("snapshots", AgentRunSnapshot.Type).
			Ref("agent_run"),
	}
}

func (AgentRun) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "owner_user_id", "created_at"),
		index.Fields("tenant_id", "workflow", "created_at"),
	}
}

type AgentRunSubject struct {
	ent.Schema
}

func (AgentRunSubject) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (AgentRunSubject) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Enum("subject_kind").Values("domain", "external"),
		field.String("entity_kind"),
		field.UUID("domain_entity_id", uuid.UUID{}).Optional().Nillable(),
		field.String("external_entity_id").Optional().Nillable(),
		field.JSON("metadata", map[string]any{}).
			SchemaType(schemaTypeJsonB).
			Optional(),
	}
}

func (AgentRunSubject) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "subject_kind", "entity_kind"),
		index.Fields("tenant_id", "domain_entity_id"),
		index.Fields("tenant_id", "external_entity_id"),
	}
}

type AgentRunSnapshot struct {
	ent.Schema
}

func (AgentRunSnapshot) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentRunSnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("agent_run_id", uuid.UUID{}),
		field.UUID("parent_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("status").Values("pending", "completed", "aborted", "failed"),
		field.String("finish_reason"),
		field.Time("heartbeat_at").Optional().Nillable(),
		field.Bytes("state").Nillable(),
		field.Bytes("error").Optional().Nillable(),
	}
}

func (AgentRunSnapshot) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("agent_run", AgentRun.Type).
			Required().
			Unique().
			Field("agent_run_id"),

		edge.To("parent", AgentRunSnapshot.Type).
			Unique().
			Field("parent_id"),
	}
}

func (AgentRunSnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_run_id"),
		index.Fields("tenant_id", "agent_run_id", "created_at"),
	}
}

type AgentRunResult struct {
	ent.Schema
}

func (AgentRunResult) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentRunResult) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("agent_run_id", uuid.UUID{}),
		field.Bytes("output"),
	}
}

func (AgentRunResult) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("agent_run", AgentRunSnapshot.Type).
			Required().
			Unique().
			Field("agent_run_id"),
		edge.From("findings", AgentRunFinding.Type).
			Ref("agent_run_result"),
	}
}

func (AgentRunResult) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_run_id").Unique(),
	}
}

type AgentRunFinding struct {
	ent.Schema
}

func (AgentRunFinding) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentRunFinding) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("agent_run_result_id", uuid.UUID{}),
		field.String("finding_kind").NotEmpty(),
		field.Text("content").NotEmpty(),
	}
}

func (AgentRunFinding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("agent_run_result", AgentRunResult.Type).
			Required().
			Unique().
			Field("agent_run_result_id"),

		edge.To("citations", AgentRunCitation.Type).
			Through("finding_citations", AgentRunFindingCitation.Type),
	}
}

func (AgentRunFinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_run_result_id"),
	}
}

type AgentRunCitation struct {
	ent.Schema
}

func (AgentRunCitation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentRunCitation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("kind").NotEmpty(),
		field.Text("summary").NotEmpty(),
		field.UUID("knowledge_entity_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("knowledge_relationship_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("knowledge_evidence_id", uuid.UUID{}).Optional().Nillable(),
		field.String("domain_entity_type").Optional(),
		field.UUID("domain_entity_id", uuid.UUID{}).Optional().Nillable(),
		field.JSON("domain_entity_snapshot", map[string]any{}).
			SchemaType(schemaTypeJsonB).
			Optional(),
	}
}

func (AgentRunCitation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("knowledge_entity", KnowledgeEntity.Type).
			Unique().
			Field("knowledge_entity_id"),
		edge.To("knowledge_relationship", KnowledgeRelationship.Type).
			Unique().
			Field("knowledge_relationship_id"),
		edge.To("knowledge_evidence", KnowledgeEvidence.Type).
			Unique().
			Field("knowledge_evidence_id"),

		edge.From("findings", AgentRunFinding.Type).
			Through("finding_citations", AgentRunFindingCitation.Type).
			Ref("citations"),
	}
}

func (AgentRunCitation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "kind"),
		index.Fields("tenant_id", "domain_entity_type", "domain_entity_id"),
		index.Fields("tenant_id", "knowledge_entity_id"),
		index.Fields("tenant_id", "knowledge_relationship_id"),
		index.Fields("tenant_id", "knowledge_evidence_id"),
	}
}

type AgentRunFindingCitation struct {
	ent.Schema
}

func (AgentRunFindingCitation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentRunFindingCitation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("finding_id", uuid.UUID{}),
		field.UUID("citation_id", uuid.UUID{}),
		field.String("support_kind").NotEmpty(),
	}
}

func (AgentRunFindingCitation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("finding", AgentRunFinding.Type).
			Required().
			Unique().
			Field("finding_id"),
		edge.To("citation", AgentRunCitation.Type).
			Required().
			Unique().
			Field("citation_id"),
	}
}

func (AgentRunFindingCitation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "finding_id"),
		index.Fields("tenant_id", "citation_id"),
	}
}
