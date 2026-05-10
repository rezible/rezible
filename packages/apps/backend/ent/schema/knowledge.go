package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type KnowledgeFact struct {
	ent.Schema
}

func (KnowledgeFact) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (KnowledgeFact) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("kind").NotEmpty(),
		field.String("display_name").NotEmpty(),
		field.Text("description").Optional(),
		field.JSON("properties", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
	}
}

func (KnowledgeFact) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("aliases", KnowledgeFactAlias.Type).
			Ref("fact"),
		edge.From("source_relationships", KnowledgeFactRelationship.Type).
			Ref("source_fact"),
		edge.From("target_relationships", KnowledgeFactRelationship.Type).
			Ref("target_fact"),
	}
}

func (KnowledgeFact) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "kind"),
		index.Fields("tenant_id", "updated_at"),
	}
}

type KnowledgeFactAlias struct {
	ent.Schema
}

func (KnowledgeFactAlias) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (KnowledgeFactAlias) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("fact_id", uuid.UUID{}),
		field.String("display_name").Optional(),
		field.String("provider"),
		field.String("provider_source"),
		field.String("provider_subject_ref"),
	}
}

func (KnowledgeFactAlias) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("fact", KnowledgeFact.Type).
			Required().Unique().
			Field("fact_id"),
		edge.From("provenance", KnowledgeFactProvenance.Type).
			Ref("alias"),
	}
}

func (KnowledgeFactAlias) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "fact_id"),
	}
}

type KnowledgeFactRelationship struct {
	ent.Schema
}

func (KnowledgeFactRelationship) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (KnowledgeFactRelationship) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("source_fact_id", uuid.UUID{}),
		field.UUID("target_fact_id", uuid.UUID{}),
		field.String("kind").NotEmpty(),
		field.String("display_name").Optional(),
		field.Text("description").Optional(),
		field.JSON("properties", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
	}
}

func (KnowledgeFactRelationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("source_fact", KnowledgeFact.Type).
			Required().
			Unique().
			Field("source_fact_id"),
		edge.To("target_fact", KnowledgeFact.Type).
			Required().
			Unique().
			Field("target_fact_id"),

		edge.From("provenance", KnowledgeFactProvenance.Type).
			Ref("relationship"),
	}
}

func (KnowledgeFactRelationship) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "kind").
			Edges("source_fact", "target_fact").
			Unique(),
		index.Fields("tenant_id", "kind"),
		index.Fields("tenant_id", "source_fact_id"),
		index.Fields("tenant_id", "target_fact_id"),
		index.Fields("tenant_id", "updated_at"),
	}
}

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
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("alias_id", uuid.UUID{}).Optional().Nillable().
			Comment("Alias this provenance supports. Exactly one of alias_id or relationship_id must be set."),
		field.UUID("relationship_id", uuid.UUID{}).Optional().Nillable().
			Comment("Relationship this provenance supports. Exactly one of alias_id or relationship_id must be set."),
		field.UUID("normalized_event_id", uuid.UUID{}).
			Comment("Normalized event that produced this provenance record"),
		field.String("source").NotEmpty(),
	}
}

func (KnowledgeFactProvenance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("alias", KnowledgeFactAlias.Type).
			Unique().
			Field("alias_id"),

		edge.To("relationship", KnowledgeFactRelationship.Type).
			Unique().
			Field("relationship_id"),

		edge.To("normalized_event", NormalizedEvent.Type).
			Unique().Required().
			Field("normalized_event_id"),
	}
}

func (KnowledgeFactProvenance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "alias_id"),
		index.Fields("tenant_id", "relationship_id"),
		index.Fields("tenant_id", "normalized_event_id"),
	}
}
