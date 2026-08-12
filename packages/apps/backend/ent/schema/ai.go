package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
)

type AgentSession struct {
	ent.Schema
}

func (AgentSession) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentSession) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("agent_name").NotEmpty(),
		field.UUID("owner_user_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Strings("scopes").Default([]string{}),
		field.Bytes("input"),
		field.JSON("metadata", map[string]any{}).
			SchemaType(schemaTypeJsonB).
			Optional(),
	}
}

func (AgentSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("owner_user", User.Type).
			Unique().
			Field("owner_user_id"),
		edge.To("turns", AgentTurn.Type),
		edge.To("messages", AgentMessage.Type),
		edge.To("artifacts", AgentArtifact.Type),
	}
}

func (AgentSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "owner_user_id", "created_at"),
		index.Fields("tenant_id", "agent_name", "created_at"),
	}
}

type AgentTurn struct {
	ent.Schema
}

func (AgentTurn) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentTurn) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}),
		field.UUID("agent_session_id", uuid.UUID{}).Immutable(),
		field.Int("sequence").Immutable().Positive(),
		field.Int64("river_job_id"),
		field.JSON("input_tool_resume", &aix.ToolResume{}).
			Optional().
			SchemaType(schemaTypeJsonB),
		field.UUID("input_message_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Enum("status").Values("queued", "running", "completed", "failed", "aborted"),
		field.Time("started_at").Optional().Nillable(),
		field.Time("finished_at").Optional().Nillable(),
		field.String("finish_reason").Default(""),
		field.String("error").Optional().Nillable(),
	}
}

func (AgentTurn) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("agent_session", AgentSession.Type).
			Ref("turns").
			Required().
			Unique().
			Immutable().
			Field("agent_session_id"),

		edge.To("input_message", AgentMessage.Type).
			Unique().
			Field("input_message_id"),

		edge.To("messages", AgentMessage.Type),
		edge.To("artifacts", AgentArtifact.Type),
		edge.To("knowledge_citations", AgentTurnKnowledgeCitation.Type),
	}
}

func (AgentTurn) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("agent_session_id", "sequence").
			Unique(),

		index.Fields("agent_session_id").
			Unique().
			StorageKey("agent_turn_one_active_per_session").
			Annotations(entsql.IndexWhere("status IN ('queued', 'running')")),

		index.Fields("input_message_id").
			Unique().
			StorageKey("agent_turn_input_message_unique").
			Annotations(entsql.IndexWhere("input_message_id IS NOT NULL")),

		index.Fields("tenant_id", "agent_session_id", "created_at"),
	}
}

type AgentMessage struct {
	ent.Schema
}

func (AgentMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentMessage) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("agent_session_id", uuid.UUID{}).Immutable(),
		field.UUID("agent_turn_id", uuid.UUID{}).Immutable(),
		field.Int("sequence").Positive().Immutable(),
		field.Enum("role").Values("user", "model", "tool", "system"),
		field.JSON("content", []*ai.Part{}).
			SchemaType(schemaTypeJsonB),
		field.JSON("metadata", map[string]any{}).
			Default(map[string]any{}).
			SchemaType(schemaTypeJsonB),
		field.Bool("visible").
			Default(true),
	}
}

func (AgentMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("agent_session", AgentSession.Type).
			Ref("messages").
			Required().
			Unique().
			Immutable().
			Field("agent_session_id"),

		edge.From("agent_turn", AgentTurn.Type).
			Ref("messages").
			Required().
			Unique().
			Immutable().
			Field("agent_turn_id"),
	}
}

func (AgentMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("agent_session_id", "sequence").Unique(),
		index.Fields("tenant_id", "agent_session_id", "created_at"),
		index.Fields("tenant_id", "agent_turn_id", "sequence"),
	}
}

type AgentArtifact struct {
	ent.Schema
}

func (AgentArtifact) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentArtifact) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("agent_session_id", uuid.UUID{}).Immutable(),
		field.UUID("last_agent_turn_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("name").
			NotEmpty(),
		field.JSON("parts", []*ai.Part{}).
			SchemaType(schemaTypeJsonB),
		field.JSON("metadata", map[string]any{}).
			Default(map[string]any{}).
			SchemaType(schemaTypeJsonB),
	}
}

func (AgentArtifact) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("agent_session", AgentSession.Type).
			Ref("artifacts").
			Required().
			Unique().
			Immutable().
			Field("agent_session_id"),

		edge.From("agent_turn", AgentTurn.Type).
			Ref("artifacts").
			Unique().
			Field("last_agent_turn_id"),
	}
}

func (AgentArtifact) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("agent_session_id", "name").Unique(),
		index.Fields("tenant_id", "agent_session_id", "created_at"),
		index.Fields("tenant_id", "last_agent_turn_id"),
	}
}

type AgentTurnKnowledgeCitation struct {
	ent.Schema
}

func (AgentTurnKnowledgeCitation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentTurnKnowledgeCitation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("agent_turn_id", uuid.UUID{}),
		field.UUID("knowledge_evidence_id", uuid.UUID{}),
		field.Text("summary").NotEmpty(),
	}
}

func (AgentTurnKnowledgeCitation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("agent_turn", AgentTurn.Type).
			Ref("knowledge_citations").
			Unique().
			Required().
			Field("agent_turn_id"),

		edge.To("knowledge_evidence", KnowledgeEvidence.Type).
			Unique().
			Required().
			Field("knowledge_evidence_id"),
	}
}

func (AgentTurnKnowledgeCitation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_turn_id", "knowledge_evidence_id").Unique(),
		index.Fields("tenant_id", "knowledge_evidence_id"),
	}
}
