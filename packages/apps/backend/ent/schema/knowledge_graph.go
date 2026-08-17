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
		"actor",          // Human or organizational participant; e.g. customer, on-call team.
		"system",         // Top-level software or business system; e.g. ecommerce platform, billing system.
		"container",      // Independently deployable or runnable unit; e.g. service, database.
		"component",      // Internal part of a container/system; e.g. module, background worker.
		"infrastructure", // Runtime or platform resource; e.g. Kubernetes cluster, namespace.
		"code",           // Source-code artifact; e.g. repository, package.
		"process",        // Business or operational workflow; e.g. checkout flow, deploy process.
		"domain_object",  // Business/domain object; e.g. customer, order.
		"concern",        // External force, risk, or requirement; e.g. compliance rule, latency target.
		"decision",       // Chosen design or operating tradeoff; e.g. ADR, mitigation choice.
		"event",          // Time-bound occurrence; e.g. incident, deploy.
		"signal",         // Observable telemetry or notification; e.g. alert, metric.
	}

	knowledgeRelationshipKinds = []string{
		"contains",        // Structural containment; e.g. system contains service or namespace contains pod.
		"interacts_with",  // Runtime or logical interaction; e.g. service calls API or worker reads queue.
		"depends_on",      // Required dependency; e.g. service depends on database or provider.
		"runs_on",         // Runtime placement; e.g. service runs on cluster or pod runs on node.
		"owns",            // Accountability or stewardship; e.g. team owns service or group owns process.
		"supports",        // Capability or dependency support; e.g. service supports checkout flow.
		"participates_in", // Actor or object participation; e.g. user in team or service in process.
		"controls",        // Control exerted over another subject; e.g. runbook controls recovery or rate limit controls API.
		"observes",        // Telemetry or observation path; e.g. alert observes service or dashboard observes queue.
		"influences",      // Non-binding causal pressure; e.g. regulation influences decision.
		"constrains",      // Hard limit or rule; e.g. SLO constrains design or policy constrains access.
		"addresses",       // Response to concern or risk; e.g. decision addresses hazard.
		"impacts",         // Effect or consequence; e.g. incident impacts customer or deploy impacts service.
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
