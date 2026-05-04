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
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (KnowledgeEntity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("aliases", KnowledgeEntityAlias.Type).Ref("entity"),
		edge.From("source_relationships", KnowledgeRelationship.Type).Ref("source_entity"),
		edge.From("target_relationships", KnowledgeRelationship.Type).Ref("target_entity"),
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
	}
}

func (KnowledgeEntityAlias) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("entity_id", uuid.UUID{}),
		field.String("provider").NotEmpty(),
		field.String("source").NotEmpty(),
		field.String("external_kind").NotEmpty(),
		field.String("external_id").NotEmpty(),
		field.UUID("normalized_event_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("first_seen_at").Default(time.Now),
		field.Time("last_seen_at").Default(time.Now),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
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
		edge.From("provenance", KnowledgeFactProvenance.Type).Ref("alias"),
	}
}

func (KnowledgeEntityAlias) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "provider", "source", "external_kind", "external_id").
			Unique(),
		index.Fields("tenant_id", "entity_id"),
	}
}
