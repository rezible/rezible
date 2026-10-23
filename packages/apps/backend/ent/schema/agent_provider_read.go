package schema

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent/schema/schematypes"
)

// AgentProviderRead is one read an agent made of a connected provider's live data, with the native query that
// ran. A turn's provider reads are its read budget, counted across attempts.
type AgentProviderRead struct {
	ent.Schema
}

func (AgentProviderRead) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (AgentProviderRead) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("agent_session_id", uuid.UUID{}).Immutable(),
		field.UUID("agent_turn_id", uuid.UUID{}).Immutable(),
		field.String("tool_name").NotEmpty().Immutable(),
		field.UUID("service_entity_id", uuid.UUID{}),
		// IntegrationID is the installation that ran the read; nil once it is deleted.
		field.UUID("integration_id", uuid.UUID{}).Optional().Nillable(),
		// Input is the tool's input as the agent sent it.
		field.JSON("input", json.RawMessage{}).
			SchemaType(schemaTypeJsonB).
			Immutable(),
		field.Enum("status").
			Values("running", "succeeded", "failed").
			Default("running"),
		// Read describes the read that ran: the native query, the range and step, limits and the Explore URL.
		field.JSON("read", json.RawMessage{}).
			SchemaType(schemaTypeJsonB).
			Optional(),
		// Highlight is up to three of the read's items the investigation cited, newest citation first.
		field.JSON("highlight", []schematypes.ProviderReadItem{}).
			SchemaType(schemaTypeJsonB).
			Optional(),
		field.String("error").Optional().Nillable(),
		field.Time("finished_at").Optional().Nillable(),
	}
}

func (AgentProviderRead) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("agent_session", AgentSession.Type).
			Required().
			Unique().
			Immutable().
			Field("agent_session_id"),
		edge.From("agent_turn", AgentTurn.Type).
			Ref("provider_reads").
			Required().
			Unique().
			Immutable().
			Field("agent_turn_id"),
		edge.To("service_entity", KnowledgeEntity.Type).
			Required().
			Unique().
			Field("service_entity_id").
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.To("integration", Integration.Type).
			Unique().
			Field("integration_id").
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (AgentProviderRead) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("agent_turn_id"),
	}
}
