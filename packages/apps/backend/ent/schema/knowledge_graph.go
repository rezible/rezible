package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/schema/schematypes"
)

var (
	knowledgeEntityKinds = []string{
		"actor",
		"system",
		"container",
		"component",
		"code",
		"deployment_node",
		"process",
		"concern",
		"decision",
		"event",
	}

	knowledgeRelationshipKinds = []string{
		"contains",
		"interacts_with",
		"owns",
		"implemented_by",
		"runs_on",
		"control_action",
		"feedback",
		"supports",
		"participates_in",
		"influences",
		"constrains",
		"addresses",
		"impacts",
	}
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
		field.Enum("kind").
			Values(knowledgeEntityKinds...).
			Comment("Stable semantic category used by graph queries, generated views, and agent reasoning."),
		field.String("subkind").NotEmpty().
			Comment("Provider or domain subtype used for filtering, legends, and display; not product control flow."),
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
		index.Fields("tenant_id", "kind", "subkind"),
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
		field.Enum("kind").
			Values(knowledgeRelationshipKinds...).
			Immutable(),
		field.String("subkind").NotEmpty().
			Comment("Provider or domain subtype").
			Immutable(),
		field.UUID("source_entity_id", uuid.UUID{}).Immutable(),
		field.UUID("target_entity_id", uuid.UUID{}).Immutable(),
	}
}

func (KnowledgeRelationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("source_entity", KnowledgeEntity.Type).
			Required().
			Unique().
			Immutable().
			Field("source_entity_id"),
		edge.To("target_entity", KnowledgeEntity.Type).
			Required().
			Unique().
			Immutable().
			Field("target_entity_id"),
		edge.From("aliases", KnowledgeSubjectAlias.Type).
			Ref("relationship"),
	}
}

func (KnowledgeRelationship) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "kind", "subkind", "source_entity_id", "target_entity_id").Unique(),
		index.Fields("tenant_id", "source_entity_id"),
		index.Fields("tenant_id", "target_entity_id"),
		index.Fields("tenant_id", "kind", "subkind"),
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

		field.Enum("subject_kind").
			Values("entity", "relationship").
			Immutable(),

		field.String("provider").NotEmpty().Immutable(),
		field.String("provider_source").NotEmpty().Immutable(),
		field.String("provider_subject_ref").NotEmpty().Immutable(),

		field.UUID("entity_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("relationship_id", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

func (KnowledgeSubjectAlias) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields(
			"tenant_id",
			"subject_kind",
			"provider",
			"provider_source",
			"provider_subject_ref",
		).Unique(),
		index.Fields("tenant_id", "entity_id"),
		index.Fields("tenant_id", "relationship_id"),
	}
}

/* TODO: annotation/db check
CHECK (
    (
      subject_kind = 'entity'
      AND entity_id IS NOT NULL
      AND relationship_id IS NULL
    )
    OR
    (
      subject_kind = 'relationship'
      AND relationship_id IS NOT NULL
      AND entity_id IS NULL
    )
  )
*/

func (KnowledgeSubjectAlias) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entity", KnowledgeEntity.Type).
			Unique().Field("entity_id"),
		edge.To("relationship", KnowledgeRelationship.Type).
			Unique().Field("relationship_id"),

		edge.From("evidence", KnowledgeEvidence.Type).
			Ref("subject_alias"),
	}
}

type KnowledgeEvidence struct {
	ent.Schema
}

func (KnowledgeEvidence) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (KnowledgeEvidence) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),

		field.UUID("event_id", uuid.UUID{}).Immutable().
			Comment("Normalized event that produced this evidence record."),
		field.UUID("subject_alias_id", uuid.UUID{}).Immutable().
			Comment("Alias used to resolve a single entity or relationship."),

		field.Enum("kind").Immutable().
			Values("observed", "deleted").
			Comment("How this event affects evidence for the assertion."),

		field.String("assertion").NotEmpty().Immutable().
			Comment("Domain assertion supported by this evidence (eg service_exists, team_owns_service)"),

		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("effective_at").Immutable().
			Comment("Domain effective time (may differ from the event occurred_at)"),

		field.JSON("subject_state", schematypes.KnowledgeGraphSubjectState{}).
			SchemaType(schemaTypeJsonB).
			Immutable(),
	}
}

func (KnowledgeEvidence) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", NormalizedEvent.Type).
			Unique().Required().Immutable().
			Field("event_id"),
		edge.To("subject_alias", KnowledgeSubjectAlias.Type).
			Unique().Required().Immutable().
			Field("subject_alias_id"),
	}
}

func (KnowledgeEvidence) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "event_id", "subject_alias_id").Unique(),
		index.Fields("tenant_id", "subject_alias_id", "effective_at"),
		index.Fields("tenant_id", "event_id"),
	}
}
