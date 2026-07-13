package schema

import (
	"time"

	"entgo.io/ent"
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
		field.String("reference").NotEmpty(),
		field.String("display_name").Optional(),
		field.Text("description").Optional(),
		field.JSON("live_properties", map[string]any{}).
			SchemaType(schemaTypeJsonB).
			Optional().
			Default(map[string]any{}),
	}
}

func (KnowledgeEntity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("aliases", KnowledgeSubjectAlias.Type).
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
		index.Fields("tenant_id", "kind", "reference").Unique(),
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
		field.String("kind").NotEmpty(),
		field.UUID("source_entity_id", uuid.UUID{}),
		field.UUID("target_entity_id", uuid.UUID{}),
		field.Text("description").Optional(),
		field.JSON("properties", map[string]any{}).
			Optional().
			SchemaType(schemaTypeJsonB),
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
		edge.From("aliases", KnowledgeSubjectAlias.Type).
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
	}
}

type KnowledgeSubjectAlias struct {
	ent.Schema
}

func (KnowledgeSubjectAlias) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (KnowledgeSubjectAlias) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Enum("subject_kind").Values("entity", "relationship").Immutable(),
		field.String("provider").NotEmpty().Immutable(),
		field.String("provider_subject_ref").NotEmpty().Immutable(),
		field.UUID("entity_id", uuid.UUID{}).Optional().Immutable(),
		field.UUID("relationship_id", uuid.UUID{}).Optional().Immutable(),
		field.String("description"),
		field.Time("first_observed_at").Optional().
			Default(time.Now).Immutable(),
		field.Time("last_observed_at").Optional().
			Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable().
			Comment("Time observed explicit evidence that this subject no longer exists or applies."),
	}
}

func (KnowledgeSubjectAlias) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("evidence", KnowledgeEvidence.Type).
			Ref("alias"),

		edge.To("entity", KnowledgeEntity.Type).
			Unique().Immutable().Field("entity_id"),
		edge.To("relationship", KnowledgeRelationship.Type).
			Unique().Immutable().Field("relationship_id"),
	}
}

func (KnowledgeSubjectAlias) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "entity_id"),
		index.Fields("tenant_id", "relationship_id"),
		index.Fields("tenant_id", "subject_kind", "entity_id", "relationship_id"),
		index.Fields("tenant_id", "subject_kind", "provider", "provider_subject_ref").Unique(),
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
		field.UUID("event_id", uuid.UUID{}).Immutable().
			Comment("Normalized event that produced this evidence record."),
		field.String("assertion").NotEmpty().Immutable().
			Comment("Domain assertion supported by this evidence (eg service_exists, team_owns_service)"),
		field.Enum("evidence_kind").Immutable().
			Values("observed", "changed", "deleted").
			Comment("How this event affects evidence for the assertion."),
		field.UUID("alias_id", uuid.UUID{}).Immutable().
			Comment("Alias used to resolve a single entity or relationship from evidence."),
		field.Time("effective_at").Immutable().
			Comment("Domain effective time (may differ from the event occurred_at)"),
		field.JSON("properties", map[string]any{}).SchemaType(schemaTypeJsonB).Immutable(),
	}
}

func (KnowledgeEvidence) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", NormalizedEvent.Type).
			Unique().Required().Immutable().
			Field("event_id"),
		edge.To("alias", KnowledgeSubjectAlias.Type).
			Unique().Required().Immutable().
			Field("alias_id"),
	}
}

func (KnowledgeEvidence) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "event_id", "alias_id", "evidence_kind").Unique(),
		index.Fields("tenant_id", "alias_id"),
		index.Fields("tenant_id", "event_id"),
		index.Fields("tenant_id", "effective_at"),
	}
}
