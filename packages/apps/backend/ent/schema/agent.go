package schema

import (
	"time"

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

var (
	agentRunWorkflowKinds = []string{
		"incident_context_pack",
		"alert_investigation",
		"retrospective_analysis",
	}
	agentRunStatuses = []string{"queued", "running", "succeeded", "failed"}
)

func (AgentRun) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("agent_case_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("workflow_kind").Values(agentRunWorkflowKinds...),
		field.Enum("status").Values(agentRunStatuses...).Default("queued"),
		field.String("idempotency_key").NotEmpty(),
		field.String("subject_kind").Optional(),
		field.UUID("subject_id", uuid.UUID{}).Optional().Nillable(),
		field.JSON("trigger_metadata", map[string]any{}).SchemaType(schemaTypeJsonB).
			Optional(),
		field.JSON("model_metadata", map[string]any{}).
			SchemaType(schemaTypeJsonB).
			Optional(),
		field.String("error_code").Optional(),
		field.Text("error_message").Optional(),
		field.Time("queued_at").Default(time.Now),
		field.Time("started_at").Optional().Nillable(),
		field.Time("completed_at").Optional().Nillable(),
		field.Time("failed_at").Optional().Nillable(),
	}
}

func (AgentRun) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("agent_case", AgentCase.Type).
			Unique().
			Field("agent_case_id"),
		edge.From("case_steps", AgentCaseStep.Type).
			Ref("agent_run"),
		edge.From("case_artifacts", AgentCaseArtifact.Type).
			Ref("agent_run"),
		edge.From("case_conclusions", AgentCaseConclusion.Type).
			Ref("agent_run"),
		edge.From("feedback", AgentRunFeedback.Type).
			Ref("agent_run"),
	}
}

func (AgentRun) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "workflow_kind", "idempotency_key").Unique(),
		index.Fields("tenant_id", "status", "updated_at"),
		index.Fields("tenant_id", "subject_kind", "subject_id"),
		index.Fields("tenant_id", "agent_case_id"),
	}
}

type AgentCase struct {
	ent.Schema
}

func (AgentCase) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

var (
	agentCaseStatuses  = []string{"open", "running", "completed", "failed", "cancelled"}
	agentCaseStepKinds = []string{
		"planning",
		"retrieval",
		"tool_call",
		"observation",
		"reasoning",
		"conclusion",
		"system",
	}
)

func (AgentCase) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Enum("status").Values(agentCaseStatuses...).Default("open"),
		field.String("title").NotEmpty(),
		field.Text("query").Optional(),
		field.Enum("workflow_kind").Values(agentRunWorkflowKinds...).Optional(),
		field.String("subject_kind").Optional(),
		field.UUID("subject_id", uuid.UUID{}).Optional().Nillable(),
		field.JSON("trigger_metadata", map[string]any{}).SchemaType(schemaTypeJsonB).
			Optional(),
		field.Text("summary").Optional(),
		field.String("error_code").Optional(),
		field.Text("error_message").Optional(),
	}
}

func (AgentCase) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("runs", AgentRun.Type).
			Ref("agent_case"),
		edge.From("steps", AgentCaseStep.Type).
			Ref("agent_case"),
		edge.From("artifacts", AgentCaseArtifact.Type).
			Ref("agent_case"),
		edge.From("conclusions", AgentCaseConclusion.Type).
			Ref("agent_case"),
	}
}

func (AgentCase) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "status", "updated_at"),
		index.Fields("tenant_id", "workflow_kind", "status"),
		index.Fields("tenant_id", "subject_kind", "subject_id"),
	}
}

type AgentCaseStep struct {
	ent.Schema
}

func (AgentCaseStep) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentCaseStep) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("agent_case_id", uuid.UUID{}),
		field.UUID("agent_run_id", uuid.UUID{}).Optional().Nillable(),
		field.Int("sequence"),
		field.Enum("kind").Values(agentCaseStepKinds...),
		field.String("title").NotEmpty(),
		field.Text("summary").Optional(),
		field.JSON("input", map[string]any{}).SchemaType(schemaTypeJsonB).
			Optional(),
		field.JSON("output", map[string]any{}).SchemaType(schemaTypeJsonB).
			Optional(),
		field.Time("started_at").Optional().Nillable(),
		field.Time("completed_at").Optional().Nillable(),
	}
}

func (AgentCaseStep) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("agent_case", AgentCase.Type).
			Required().
			Unique().
			Field("agent_case_id"),
		edge.To("agent_run", AgentRun.Type).
			Unique().
			Field("agent_run_id"),
		edge.From("artifacts", AgentCaseArtifact.Type).
			Ref("agent_case_step"),
		edge.From("conclusions", AgentCaseConclusion.Type).
			Ref("agent_case_step"),
	}
}

func (AgentCaseStep) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_case_id", "sequence").Unique(),
		index.Fields("tenant_id", "agent_case_id", "kind"),
		index.Fields("tenant_id", "agent_run_id"),
	}
}

type AgentCaseArtifact struct {
	ent.Schema
}

func (AgentCaseArtifact) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

var (
	agentCaseArtifactKinds = []string{
		"entity_ref",
		"relationship_ref",
		"evidence_ref",
		"reference_ref",
		"context",
		"model",
		"custom",
	}
)

func (AgentCaseArtifact) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("agent_case_id", uuid.UUID{}),
		field.UUID("agent_case_step_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("agent_run_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("kind").Values(agentCaseArtifactKinds...),
		field.String("role").Optional(),
		field.String("name").NotEmpty(),
		field.JSON("payload", map[string]any{}).SchemaType(schemaTypeJsonB).
			Optional(),
		field.Bool("redacted").Default(false),
	}
}

func (AgentCaseArtifact) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("agent_case", AgentCase.Type).
			Required().
			Unique().
			Field("agent_case_id"),
		edge.To("agent_case_step", AgentCaseStep.Type).
			Unique().
			Field("agent_case_step_id"),
		edge.To("agent_run", AgentRun.Type).
			Unique().
			Field("agent_run_id"),
	}
}

func (AgentCaseArtifact) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_case_id", "kind"),
		index.Fields("tenant_id", "agent_case_id", "role"),
		index.Fields("tenant_id", "agent_case_id", "name"),
		index.Fields("tenant_id", "agent_case_step_id"),
		index.Fields("tenant_id", "agent_run_id"),
	}
}

type AgentCaseConclusion struct {
	ent.Schema
}

func (AgentCaseConclusion) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentCaseConclusion) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("agent_case_id", uuid.UUID{}),
		field.UUID("agent_case_step_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("agent_run_id", uuid.UUID{}).Optional().Nillable(),
		field.String("kind").NotEmpty(),
		field.Text("summary").Optional(),
		field.String("confidence").Optional(),
		field.Strings("recommended_actions").Optional(),
		field.Strings("limitations").Optional(),
		field.JSON("payload", map[string]any{}).SchemaType(schemaTypeJsonB).
			Optional(),
	}
}

func (AgentCaseConclusion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("agent_case", AgentCase.Type).
			Required().
			Unique().
			Field("agent_case_id"),
		edge.To("agent_case_step", AgentCaseStep.Type).
			Unique().
			Field("agent_case_step_id"),
		edge.To("agent_run", AgentRun.Type).
			Unique().
			Field("agent_run_id"),
	}
}

func (AgentCaseConclusion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_case_id", "kind"),
		index.Fields("tenant_id", "agent_case_step_id"),
		index.Fields("tenant_id", "agent_run_id"),
	}
}

type AgentRunFeedback struct {
	ent.Schema
}

func (AgentRunFeedback) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentRunFeedback) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("agent_run_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}).Optional().Nillable(),
		field.Int("rating").Optional(),
		field.Text("comment").Optional(),
		field.JSON("properties", map[string]any{}).
			Optional().
			SchemaType(schemaTypeJsonB),
	}
}

func (AgentRunFeedback) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("agent_run", AgentRun.Type).
			Required().
			Unique().
			Field("agent_run_id"),
		edge.To("user", User.Type).
			Unique().
			Field("user_id"),
	}
}

func (AgentRunFeedback) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_run_id"),
		index.Fields("tenant_id", "user_id"),
	}
}
