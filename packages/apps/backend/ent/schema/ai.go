package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type AiAgentRun struct {
	ent.Schema
}

func (AiAgentRun) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AiAgentRun) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("agent_name").NotEmpty(),
		field.UUID("owner_user_id", uuid.UUID{}),
		field.Strings("scopes").Default([]string{}),
		field.Time("started_at").Optional().Nillable(),
		field.Bytes("input"),
		field.JSON("metadata", map[string]any{}).
			SchemaType(schemaTypeJsonB).
			Optional(),
	}
}

func (AiAgentRun) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("owner_user", User.Type).
			Required().
			Unique().
			Field("owner_user_id"),
		edge.To("result", AiAgentRunResult.Type).Unique(),
		edge.From("snapshots", AiAgentRunSnapshot.Type).
			Ref("ai_agent_run"),
	}
}

func (AiAgentRun) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "owner_user_id", "created_at"),
		index.Fields("tenant_id", "agent_name", "created_at"),
	}
}

type AiAgentRunSnapshot struct {
	ent.Schema
}

func (AiAgentRunSnapshot) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AiAgentRunSnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("ai_agent_run_id", uuid.UUID{}),
		field.UUID("parent_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("status").Values("pending", "completed", "aborted", "failed"),
		field.String("finish_reason"),
		field.Time("heartbeat_at").Optional().Nillable(),
		field.Bytes("state").Nillable(),
		field.Bytes("error").Optional().Nillable(),
	}
}

func (AiAgentRunSnapshot) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("ai_agent_run", AiAgentRun.Type).
			Required().
			Unique().
			Field("ai_agent_run_id"),

		edge.To("parent", AiAgentRunSnapshot.Type).
			Unique().
			Field("parent_id"),
	}
}

func (AiAgentRunSnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "ai_agent_run_id"),
		index.Fields("tenant_id", "ai_agent_run_id", "created_at"),
	}
}

type AiAgentRunResult struct {
	ent.Schema
}

func (AiAgentRunResult) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AiAgentRunResult) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("ai_agent_run_id", uuid.UUID{}),
		field.Bytes("output"),
	}
}

func (AiAgentRunResult) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ai_agent_run", AiAgentRun.Type).
			Ref("result").
			Unique().
			Required().
			Field("ai_agent_run_id"),
	}
}

func (AiAgentRunResult) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "ai_agent_run_id").Unique(),
	}
}

type AiAgentRunFinding struct {
	ent.Schema
}

func (AiAgentRunFinding) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AiAgentRunFinding) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("ai_agent_run_result_id", uuid.UUID{}),
		field.String("finding_kind").NotEmpty(),
		field.Text("content").NotEmpty(),
	}
}

func (AiAgentRunFinding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("ai_agent_run_result", AiAgentRunResult.Type).
			Required().
			Unique().
			Field("ai_agent_run_result_id"),

		edge.To("citations", AiAgentRunCitation.Type).
			Through("finding_citations", AiAgentRunFindingCitation.Type),
	}
}

func (AiAgentRunFinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "ai_agent_run_result_id"),
	}
}

type AiAgentRunCitation struct {
	ent.Schema
}

func (AiAgentRunCitation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AiAgentRunCitation) Fields() []ent.Field {
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

func (AiAgentRunCitation) Edges() []ent.Edge {
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

		edge.From("findings", AiAgentRunFinding.Type).
			Through("finding_citations", AiAgentRunFindingCitation.Type).
			Ref("citations"),
	}
}

func (AiAgentRunCitation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "kind"),
		index.Fields("tenant_id", "domain_entity_type", "domain_entity_id"),
		index.Fields("tenant_id", "knowledge_entity_id"),
		index.Fields("tenant_id", "knowledge_relationship_id"),
		index.Fields("tenant_id", "knowledge_evidence_id"),
	}
}

type AiAgentRunFindingCitation struct {
	ent.Schema
}

func (AiAgentRunFindingCitation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AiAgentRunFindingCitation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("finding_id", uuid.UUID{}),
		field.UUID("citation_id", uuid.UUID{}),
		field.String("support_kind").NotEmpty(),
	}
}

func (AiAgentRunFindingCitation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("finding", AiAgentRunFinding.Type).
			Required().
			Unique().
			Field("finding_id"),
		edge.To("citation", AiAgentRunCitation.Type).
			Required().
			Unique().
			Field("citation_id"),
	}
}

func (AiAgentRunFindingCitation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "finding_id"),
		index.Fields("tenant_id", "citation_id"),
	}
}
