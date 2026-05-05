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

type KnowledgeRelationship struct {
	ent.Schema
}

func (KnowledgeRelationship) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
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
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
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
		edge.From("provenance", KnowledgeFactProvenance.Type).Ref("relationship"),
	}
}

func (KnowledgeRelationship) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "source_entity_id", "target_entity_id", "kind").
			Unique(),
		index.Fields("tenant_id", "kind"),
	}
}

type KnowledgeFactProvenance struct {
	ent.Schema
}

func (KnowledgeFactProvenance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (KnowledgeFactProvenance) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("alias_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("relationship_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("normalized_event_id", uuid.UUID{}).Optional().Nillable(),
		field.String("source_provider").NotEmpty(),
		field.String("source").NotEmpty(),
		field.String("source_ref").Optional(),
		field.String("extraction_method").NotEmpty(),
		field.Float("confidence").Default(1),
		field.Time("first_seen_at").Default(time.Now),
		field.Time("last_seen_at").Default(time.Now),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
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
		index.Fields("tenant_id", "source_provider", "source", "source_ref"),
		index.Fields("tenant_id", "alias_id", "source_provider", "source", "source_ref", "extraction_method").
			Unique().
			StorageKey("knowledgefactprovenance_alias_source_unique"),
		index.Fields("tenant_id", "relationship_id", "source_provider", "source", "source_ref", "extraction_method").
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
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("fact_kind").NotEmpty(),
		field.UUID("alias_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("relationship_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("normalized_event_id", uuid.UUID{}).Optional().Nillable(),
		field.String("event_kind").NotEmpty(),
		field.String("history_key").NotEmpty(),
		field.Time("occurred_at"),
		field.Time("recorded_at").Default(time.Now),
		field.String("source_provider").NotEmpty(),
		field.String("source").NotEmpty(),
		field.String("source_ref").Optional(),
		field.String("extraction_method").NotEmpty(),
		field.JSON("attributes", map[string]any{}).
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
