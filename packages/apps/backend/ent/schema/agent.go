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
		edge.From("artifacts", AgentRunArtifact.Type).
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
	}
}

type AgentRunArtifact struct {
	ent.Schema
}

func (AgentRunArtifact) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

var (
	agentRunArtifactKinds = []string{"context", "result", "tool", "model"}
)

func (AgentRunArtifact) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("agent_run_id", uuid.UUID{}),
		field.Enum("kind").Values(agentRunArtifactKinds...),
		field.String("name").NotEmpty(),
		field.JSON("payload", map[string]any{}).SchemaType(schemaTypeJsonB).
			Optional(),
		field.Bool("redacted").Default(false),
	}
}

func (AgentRunArtifact) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("agent_run", AgentRun.Type).
			Required().
			Unique().
			Field("agent_run_id"),
	}
}

func (AgentRunArtifact) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_run_id", "kind"),
		index.Fields("tenant_id", "agent_run_id", "name"),
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
