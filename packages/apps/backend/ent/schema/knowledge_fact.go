package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type KnowledgeFactProvenance struct {
	ent.Schema
}

func (KnowledgeFactProvenance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (KnowledgeFactProvenance) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).
			Comment("Internal identifier for this knowledge fact provenance record."),
		field.UUID("alias_id", uuid.UUID{}).Optional().Nillable().
			Comment("Alias this provenance supports. Exactly one of alias_id or relationship_id must be set."),
		field.UUID("relationship_id", uuid.UUID{}).Optional().Nillable().
			Comment("Relationship this provenance supports. Exactly one of alias_id or relationship_id must be set."),
		field.UUID("normalized_event_id", uuid.UUID{}).Optional().Nillable().
			Comment("Normalized event that produced this provenance record, when available."),
		field.String("provider").NotEmpty().
			Comment("Integration provider that supplied the evidence for this fact."),
		field.String("provider_source").NotEmpty().
			Comment("Provider-specific stream, API, or dataset where the evidence was observed."),
		field.String("provider_event_ref").NotEmpty().
			Comment("Stable provider reference for the event or record that supports this fact."),
		field.String("extraction_method").NotEmpty().
			Comment("Projection, sync, or extraction method that created this provenance record."),
		field.Time("first_seen_at").Default(time.Now).
			Comment("First time this fact was observed from this provider evidence."),
		field.Time("last_seen_at").Default(time.Now).
			Comment("Most recent time this fact was observed from this provider evidence."),
	}
}

func (KnowledgeFactProvenance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("alias", KnowledgeEntityAlias.Type).
			Unique().
			Field("alias_id"),

		edge.To("relationship", KnowledgeRelationship.Type).
			Unique().
			Field("relationship_id"),

		edge.To("normalized_event", NormalizedEvent.Type).
			Unique().
			Field("normalized_event_id"),
	}
}

func (KnowledgeFactProvenance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "alias_id"),
		index.Fields("tenant_id", "relationship_id"),
		index.Fields("tenant_id", "provider", "provider_source", "provider_event_ref"),
		index.Fields("tenant_id", "alias_id", "provider", "provider_source", "provider_event_ref", "extraction_method").
			Unique().
			StorageKey("knowledgefactprovenance_alias_source_unique"),
		index.Fields("tenant_id", "relationship_id", "provider", "provider_source", "provider_event_ref", "extraction_method").
			Unique().
			StorageKey("knowledgefactprovenance_relationship_source_unique"),
	}
}

type KnowledgeFactHistory struct {
	ent.Schema
}

func (KnowledgeFactHistory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (KnowledgeFactHistory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).
			Comment("Internal identifier for this knowledge fact history record."),
		field.Enum("fact_kind").Values("alias", "relationship").Optional().Nillable().
			Comment("Kind of knowledge fact captured by this history record."),
		field.UUID("alias_id", uuid.UUID{}).Optional().Nillable().
			Comment("Alias this history record is about, when fact_kind is alias."),
		field.UUID("relationship_id", uuid.UUID{}).Optional().Nillable().
			Comment("Relationship this history record is about, when fact_kind is relationship."),
		field.UUID("normalized_event_id", uuid.UUID{}).Optional().Nillable().
			Comment("Normalized event that produced this history record, when available."),
		field.String("event_kind").NotEmpty().
			Comment("Type of history event, such as alias_observed or relationship_observed."),
		field.String("history_key").NotEmpty().
			Comment("Stable idempotency key for this history record."),
		field.Time("occurred_at").
			Comment("Time the underlying provider event or observation occurred."),
		field.Time("recorded_at").Default(time.Now).
			Comment("Time this history record was persisted."),
		field.String("provider").NotEmpty().
			Comment("Integration provider that supplied the evidence for this history record."),
		field.String("provider_source").NotEmpty().
			Comment("Provider-specific stream, API, or dataset where the evidence was observed."),
		field.String("provider_event_ref").NotEmpty().
			Comment("Stable provider reference for the event or record that supports this history record."),
		field.String("extraction_method").NotEmpty().
			Comment("Projection, sync, or extraction method that created this history record."),
		field.JSON("attributes", map[string]any{}).
			Comment("Structured details captured for this history event.").
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
	}
}

func (KnowledgeFactHistory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("alias", KnowledgeEntityAlias.Type).
			Unique().
			Field("alias_id"),
		edge.To("relationship", KnowledgeRelationship.Type).
			Unique().
			Field("relationship_id"),
		edge.To("normalized_event", NormalizedEvent.Type).
			Unique().
			Field("normalized_event_id"),
	}
}

func (KnowledgeFactHistory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "history_key").Unique(),
		index.Fields("tenant_id", "alias_id", "occurred_at"),
		index.Fields("tenant_id", "relationship_id", "occurred_at"),
		index.Fields("tenant_id", "fact_kind", "event_kind", "occurred_at"),
	}
}
