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
		field.Enum("kind").Values(
			"component",
			"service",
			"repository",
			"incident",
			"change_event",
		),
		field.String("display_name").NotEmpty(),
		field.Text("description").Optional(),
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
	}
}

func (KnowledgeEntity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "kind"),
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
		field.UUID("id", uuid.UUID{}).Default(uuid.New).
			Comment("Internal identifier for this knowledge entity alias."),
		field.UUID("entity_id", uuid.UUID{}).
			Comment("Knowledge entity this alias resolves to."),
		field.String("provider").NotEmpty().
			Comment("Integration provider that supplied this alias, such as slack or github."),
		field.String("provider_source").NotEmpty().
			Comment("Provider-specific stream, API, or dataset where this alias was observed."),
		field.String("subject_kind").NotEmpty().
			Comment("Provider-neutral type of the external subject this alias identifies."),
		field.String("subject_ref").NotEmpty().
			Comment("Stable external reference for the subject this alias identifies."),
		field.UUID("normalized_event_id", uuid.UUID{}).Optional().Nillable().
			Comment("Normalized event that most recently observed or updated this alias."),
		field.Time("first_seen_at").Default(time.Now).
			Comment("First time this alias was observed."),
		field.Time("last_seen_at").Default(time.Now).
			Comment("Most recent time this alias was observed."),
	}
}

func (KnowledgeEntityAlias) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entity", KnowledgeEntity.Type).
			Required().
			Unique().
			Field("entity_id"),
		edge.To("normalized_event", NormalizedEvent.Type).
			Unique().
			Field("normalized_event_id"),

		edge.From("provenance", KnowledgeFactProvenance.Type).
			Ref("alias"),
	}
}

func (KnowledgeEntityAlias) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "provider", "provider_source", "subject_kind", "subject_ref").
			Unique(),
		index.Fields("tenant_id", "entity_id"),
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
		field.Time("first_seen_at").Default(time.Now),
		field.Time("last_seen_at").Default(time.Now),
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

		edge.From("provenance", KnowledgeFactProvenance.Type).
			Ref("relationship"),
	}
}

func (KnowledgeRelationship) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "kind").
			Edges("source_entity", "target_entity").
			Unique(),
		index.Fields("tenant_id", "kind"),
	}
}
