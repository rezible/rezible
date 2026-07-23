package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
		field.UUID("owner_user_id", uuid.UUID{}),
		field.Strings("default_scopes").Default([]string{}),
		field.JSON("metadata", map[string]any{}).
			SchemaType(schemaTypeJsonB).
			Optional(),
	}
}

func (AgentSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("owner_user", User.Type).
			Required().
			Unique().
			Field("owner_user_id"),
		edge.To("turns", AgentTurn.Type),
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
		field.Int64("river_job_id"),
		field.UUID("parent_id", uuid.UUID{}).Optional().Nillable(),
		field.Strings("scopes").Optional(),
		field.Bytes("input"),
		field.Enum("status").Values("queued", "running", "completed", "failed", "aborted"),
		field.Time("started_at").Optional().Nillable(),
		field.Time("finished_at").Optional().Nillable(),
		field.String("finish_reason").Default(""),
		field.Bytes("state").Optional(),
		field.Bytes("error").Optional(),
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

		edge.To("parent", AgentTurn.Type).
			Unique().
			Field("parent_id"),
		edge.From("children", AgentTurn.Type).
			Ref("parent"),

		edge.To("knowledge_citations", AgentTurnKnowledgeCitation.Type),
	}
}

func (AgentTurn) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("agent_session_id").
			Unique().
			StorageKey("agent_turn_one_running_per_session").
			Annotations(entsql.IndexWhere("status = 'running'")),
		index.Fields("parent_id").
			Unique().
			StorageKey("agent_turn_one_successful_child_per_parent").
			Annotations(entsql.IndexWhere("parent_id IS NOT NULL AND status IN ('running', 'completed')")),
		index.Fields("agent_session_id").
			Unique().
			StorageKey("agent_turn_one_successful_root_per_session").
			Annotations(entsql.IndexWhere("parent_id IS NULL AND status IN ('running', 'completed')")),
		index.Fields("tenant_id", "agent_session_id", "created_at"),
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
		field.UUID("knowledge_entity_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("knowledge_relationship_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("knowledge_evidence_id", uuid.UUID{}).Optional().Nillable(),
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

func (AgentTurnKnowledgeCitation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_turn_id", "knowledge_evidence_id").Unique(),
		index.Fields("tenant_id", "knowledge_entity_id"),
		index.Fields("tenant_id", "knowledge_relationship_id"),
		index.Fields("tenant_id", "knowledge_evidence_id"),
	}
}
