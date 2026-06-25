package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type AgentTask struct {
	ent.Schema
}

func (AgentTask) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentTask) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("owner_user_id", uuid.UUID{}),
		field.String("workflow").NotEmpty(),
		field.Bytes("input").
			NotEmpty(),
		field.String("trigger_kind").
			NotEmpty(),
		field.JSON("trigger_metadata", map[string]any{}).
			SchemaType(schemaTypeJsonB).
			Optional(),
	}
}

func (AgentTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("owner_user", User.Type).
			Required().
			Unique().
			Field("owner_user_id"),
		edge.From("subjects", AgentTaskSubject.Type).
			Ref("task"),
		edge.From("runs", AgentRun.Type).
			Ref("task"),
		edge.From("citations", AgentRunCitation.Type).
			Ref("agent_task"),
	}
}

func (AgentTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "owner_user_id", "created_at"),
		index.Fields("tenant_id", "workflow", "created_at"),
	}
}

type AgentTaskSubject struct {
	ent.Schema
}

func (AgentTaskSubject) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (AgentTaskSubject) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("task_id", uuid.UUID{}),
		field.String("subject_kind").NotEmpty(),
		field.UUID("domain_entity_id", uuid.UUID{}).Optional().Nillable(),
		field.JSON("subject_properties", map[string]any{}).
			SchemaType(schemaTypeJsonB).
			Optional(),
	}
}

func (AgentTaskSubject) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("task", AgentTask.Type).
			Required().
			Unique().
			Field("task_id"),
	}
}

func (AgentTaskSubject) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "task_id"),
	}
}

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
		field.UUID("agent_task_id", uuid.UUID{}),
		field.Int("attempt").Positive(),
		field.Time("started_at").Optional().Nillable(),
		field.Time("cancelled_at").Optional().Nillable(),
	}
}

func (AgentRun) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("task", AgentTask.Type).
			Required().
			Unique().
			Field("agent_task_id"),
		edge.To("result", AgentRunResult.Type).
			Unique(),

		edge.From("citations", AgentRunCitation.Type).
			Ref("agent_run"),
		edge.From("findings", AgentRunFinding.Type).
			Ref("agent_run"),
		edge.From("tool_calls", AgentRunToolCall.Type).
			Ref("agent_run"),
	}
}

func (AgentRun) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_task_id", "attempt").Unique(),
		index.Fields("tenant_id", "started_at"),
	}
}

type AgentRunToolCall struct {
	ent.Schema
}

func (AgentRunToolCall) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentRunToolCall) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("agent_run_id", uuid.UUID{}),
		field.String("tool_id").NotEmpty(),
		field.Enum("status").
			Values("requested", "running", "succeeded", "failed", "cancelled"),
		field.JSON("tool_params", map[string]any{}).
			SchemaType(schemaTypeJsonB).
			Optional(),
		field.JSON("result", map[string]any{}).
			SchemaType(schemaTypeJsonB).
			Optional(),
		field.Text("error_message").Optional(),
		field.Time("started_at").Optional().Nillable(),
		field.Time("finished_at").Optional().Nillable(),
	}
}

func (AgentRunToolCall) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("agent_run", AgentRun.Type).
			Required().
			Unique().
			Field("agent_run_id"),
		edge.From("citations", AgentRunCitation.Type).
			Ref("agent_run_tool_call"),
	}
}

func (AgentRunToolCall) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_run_id", "created_at"),
		index.Fields("tenant_id", "tool_id", "created_at"),
		index.Fields("tenant_id", "status", "created_at"),
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
		field.UUID("agent_run_id", uuid.UUID{}),
		field.String("kind").NotEmpty(),
		field.Text("summary").NotEmpty(),
		field.UUID("knowledge_entity_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("knowledge_relationship_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("knowledge_evidence_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("agent_task_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("agent_run_tool_call_id", uuid.UUID{}).Optional().Nillable(),
		field.String("domain_entity_type").Optional(),
		field.UUID("domain_entity_id", uuid.UUID{}).Optional().Nillable(),
		field.JSON("domain_entity_snapshot", map[string]any{}).
			SchemaType(schemaTypeJsonB).
			Optional(),
	}
}

func (AgentRunCitation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("agent_run", AgentRun.Type).
			Required().
			Unique().
			Field("agent_run_id"),
		edge.To("knowledge_entity", KnowledgeEntity.Type).
			Unique().
			Field("knowledge_entity_id"),
		edge.To("knowledge_relationship", KnowledgeRelationship.Type).
			Unique().
			Field("knowledge_relationship_id"),
		edge.To("knowledge_evidence", KnowledgeEvidence.Type).
			Unique().
			Field("knowledge_evidence_id"),
		edge.To("agent_task", AgentTask.Type).
			Unique().
			Field("agent_task_id"),
		edge.To("agent_run_tool_call", AgentRunToolCall.Type).
			Unique().
			Field("agent_run_tool_call_id"),

		edge.From("findings", AgentRunFinding.Type).
			Through("finding_citations", AgentRunFindingCitation.Type).
			Ref("citations"),
	}
}

func (AgentRunCitation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_run_id", "created_at"),
		index.Fields("tenant_id", "kind"),
		index.Fields("tenant_id", "domain_entity_type", "domain_entity_id"),
		index.Fields("tenant_id", "knowledge_entity_id"),
		index.Fields("tenant_id", "knowledge_relationship_id"),
		index.Fields("tenant_id", "knowledge_evidence_id"),
		index.Fields("tenant_id", "agent_task_id"),
		index.Fields("tenant_id", "agent_run_tool_call_id"),
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
		field.UUID("agent_run_id", uuid.UUID{}),
		field.Int("sequence").Positive(),
		field.String("finding_kind").NotEmpty(),
		field.Text("content").NotEmpty(),
	}
}

func (AgentRunFinding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("agent_run", AgentRun.Type).
			Required().
			Unique().
			Field("agent_run_id"),

		edge.To("citations", AgentRunCitation.Type).
			Through("finding_citations", AgentRunFindingCitation.Type),
	}
}

func (AgentRunFinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_run_id", "sequence").Unique(),
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
		field.Text("error_message").Optional(),
	}
}

func (AgentRunResult) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("agent_run", AgentRun.Type).
			Required().
			Unique().
			Field("agent_run_id"),
	}
}

func (AgentRunResult) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_run_id").Unique(),
	}
}
