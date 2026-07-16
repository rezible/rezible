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
		field.Enum("status").Values("pending", "completed", "failed", "aborted"),
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
		edge.From("children", AiAgentRunSnapshot.Type).
			Ref("parent"),

		edge.From("knowledge_citations", AiAgentRunKnowledgeCitation.Type).
			Ref("ai_agent_run_snapshot"),
	}
}

func (AiAgentRunSnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "ai_agent_run_id"),
		index.Fields("tenant_id", "ai_agent_run_id", "created_at"),
	}
}

type AiAgentRunKnowledgeCitation struct {
	ent.Schema
}

func (AiAgentRunKnowledgeCitation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AiAgentRunKnowledgeCitation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("ai_agent_run_snapshot_id", uuid.UUID{}),
		field.UUID("knowledge_entity_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("knowledge_relationship_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("knowledge_evidence_id", uuid.UUID{}).Optional().Nillable(),
		field.Text("summary").NotEmpty(),
	}
}

func (AiAgentRunKnowledgeCitation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("ai_agent_run_snapshot", AiAgentRunSnapshot.Type).
			Unique().
			Required().
			Field("ai_agent_run_snapshot_id"),

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

func (AiAgentRunKnowledgeCitation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "knowledge_entity_id"),
		index.Fields("tenant_id", "knowledge_relationship_id"),
		index.Fields("tenant_id", "knowledge_evidence_id"),
	}
}
