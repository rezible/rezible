package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type KnowledgeEntity struct {
	ent.Schema
}

func (KnowledgeEntity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (KnowledgeEntity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("kind").NotEmpty(),
		field.String("display_name").NotEmpty(),
		field.Text("description").Optional(),
		field.Time("first_observed_at").Optional().Nillable().
			Comment("Time first observed evidence supporting this entity."),
		field.Time("last_observed_at").Optional().Nillable().
			Comment("Time most recently observed evidence supporting this entity."),
		field.Time("deleted_at").Optional().Nillable().
			Comment("Time observed explicit evidence that this entity no longer exists or applies."),
		field.JSON("properties", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
	}
}

func (KnowledgeEntity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("aliases", KnowledgeEntityAlias.Type).
			Ref("entity"),
		edge.From("source_relationships", KnowledgeRelationship.Type).
			Ref("source_entity"),
		edge.From("target_relationships", KnowledgeRelationship.Type).
			Ref("target_entity"),
		edge.From("evidence", KnowledgeEvidence.Type).
			Ref("entity"),
	}
}

func (KnowledgeEntity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "kind"),
		index.Fields("tenant_id", "updated_at"),
		index.Fields("tenant_id", "kind", "last_observed_at"),
		index.Fields("tenant_id", "kind", "deleted_at"),
	}
}

type KnowledgeEntityAlias struct {
	ent.Schema
}

func (KnowledgeEntityAlias) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (KnowledgeEntityAlias) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("entity_id", uuid.UUID{}),
		field.String("display_name").Optional(),
		field.String("provider"),
		field.String("provider_source"),
		field.String("provider_subject_ref"),
	}
}

func (KnowledgeEntityAlias) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entity", KnowledgeEntity.Type).
			Required().Unique().
			Field("entity_id"),
		edge.From("evidence", KnowledgeEvidence.Type).
			Ref("alias"),
	}
}

func (KnowledgeEntityAlias) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "entity_id"),
		index.Fields("tenant_id", "provider", "provider_source", "provider_subject_ref").Unique(),
	}
}

type KnowledgeRelationship struct {
	ent.Schema
}

func (KnowledgeRelationship) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (KnowledgeRelationship) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("source_entity_id", uuid.UUID{}),
		field.UUID("target_entity_id", uuid.UUID{}),
		field.String("kind").NotEmpty(),
		field.String("display_name").Optional(),
		field.Text("description").Optional(),
		field.Time("first_observed_at").Optional().Nillable().
			Comment("Time first observed evidence supporting this relationship."),
		field.Time("last_observed_at").Optional().Nillable().
			Comment("Time most recently observed evidence supporting this relationship."),
		field.Time("deleted_at").Optional().Nillable().
			Comment("Time observed explicit evidence that this relationship no longer exists or applies."),
		field.JSON("properties", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
	}
}

func (KnowledgeRelationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("source_entity", KnowledgeEntity.Type).
			Required().
			Unique().
			Field("source_entity_id"),
		edge.To("target_entity", KnowledgeEntity.Type).
			Required().
			Unique().
			Field("target_entity_id"),

		edge.From("evidence", KnowledgeEvidence.Type).
			Ref("relationship"),
	}
}

func (KnowledgeRelationship) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "kind").
			Edges("source_entity", "target_entity").
			Unique(),
		index.Fields("tenant_id", "kind"),
		index.Fields("tenant_id", "source_entity_id"),
		index.Fields("tenant_id", "target_entity_id"),
		index.Fields("tenant_id", "updated_at"),
		index.Fields("tenant_id", "kind", "last_observed_at"),
		index.Fields("tenant_id", "kind", "deleted_at"),
	}
}

type KnowledgeEvidence struct {
	ent.Schema
}

func (KnowledgeEvidence) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (KnowledgeEvidence) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Enum("subject_type").Values("entity", "relationship"),
		field.UUID("entity_id", uuid.UUID{}).Optional().Nillable().
			Comment("Entity this evidence supports. Exactly one of entity_id or relationship_id should be set."),
		field.UUID("relationship_id", uuid.UUID{}).Optional().Nillable().
			Comment("Relationship this evidence supports. Exactly one of entity_id or relationship_id should be set."),
		field.UUID("alias_id", uuid.UUID{}).Optional().Nillable().
			Comment("Provider alias used to resolve entity evidence, when applicable."),
		field.UUID("normalized_event_id", uuid.UUID{}).
			Comment("Normalized event that produced this evidence record."),
		field.String("assertion_kind").NotEmpty().
			Comment("Domain assertion supported by this evidence, such as code_repository_exists or team_owns_service."),
		field.Enum("evidence_kind").Values("observed", "changed", "deleted", "contradicted").
			Comment("How this event affects evidence for the assertion."),
		field.Time("observed_at").
			Comment("Time observed this evidence, usually the normalized event occurred_at."),
		field.Time("effective_at").Optional().Nillable().
			Comment("Provider/domain effective time when it differs from observed_at."),
		field.String("source").NotEmpty(),
		field.JSON("properties", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
	}
}

func (KnowledgeEvidence) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entity", KnowledgeEntity.Type).
			Unique().
			Field("entity_id"),

		edge.To("relationship", KnowledgeRelationship.Type).
			Unique().
			Field("relationship_id"),

		edge.To("alias", KnowledgeEntityAlias.Type).
			Unique().
			Field("alias_id"),

		edge.To("normalized_event", NormalizedEvent.Type).
			Unique().Required().
			Field("normalized_event_id"),
	}
}

func (KnowledgeEvidence) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "entity_id"),
		index.Fields("tenant_id", "relationship_id"),
		index.Fields("tenant_id", "alias_id"),
		index.Fields("tenant_id", "normalized_event_id"),
		index.Fields("tenant_id", "assertion_kind", "evidence_kind", "observed_at"),
		index.Fields("tenant_id", "normalized_event_id", "assertion_kind", "subject_type", "entity_id").Unique(),
		index.Fields("tenant_id", "normalized_event_id", "assertion_kind", "subject_type", "relationship_id").Unique(),
	}
}
