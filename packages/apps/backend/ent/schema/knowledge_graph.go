package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/schema/schematypes"
)

var (
	knowledgeEntityCategories = []string{
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

	knowledgeRelationshipPredicates = []string{
		"contains",        // Structural containment; e.g. system contains service or namespace contains pod.
		"interacts_with",  // Runtime or logical interaction; e.g. service calls API or worker reads queue.
		"calls",           // Request interaction; e.g. frontend calls API.
		"reads_from",      // Data read; e.g. service reads from database.
		"writes_to",       // Data write; e.g. service writes to database.
		"publishes_to",    // Message publication; e.g. service publishes to queue.
		"consumes_from",   // Message consumption; e.g. worker consumes from queue.
		"depends_on",      // Required dependency; e.g. service depends on database or provider.
		"runs_on",         // Runtime placement; e.g. service runs on cluster or pod runs on node.
		"owns",            // Accountability or stewardship; e.g. team owns service or group owns process.
		"supports",        // Capability or dependency support; e.g. service supports checkout flow.
		"participates_in", // Actor or object participation; e.g. user in team or service in process.
		"member_of",       // Membership; e.g. user is a member of team.
		"controls",        // Control exerted over another subject; e.g. runbook controls recovery or rate limit controls API.
		"observes",        // Telemetry or observation path; e.g. alert observes service or dashboard observes queue.
		"influences",      // Non-binding causal pressure; e.g. regulation influences decision.
		"constrains",      // Hard limit or rule; e.g. SLO constrains design or policy constrains access.
		"addresses",       // Response to concern or risk; e.g. decision addresses hazard.
		"impacts",         // Effect or consequence; e.g. incident impacts customer or deploy impacts service.
		"touches",         // Change contact; e.g. code change touches repository.
		"uses",            // General use when a more precise interaction predicate is unavailable.
		"processes",       // Domain processing; e.g. service processes order.
		"indexes",         // Indexing; e.g. worker indexes product.
		"stores",          // Storage; e.g. database stores customer.
		"indicates",       // Evidence signal; e.g. alert episode indicates operational situation.
		"classified_as",   // Classification; e.g. operational situation is classified as hazard.
		"responds_to",     // Response linkage; e.g. incident responds to operational situation.
		"mitigates",       // Risk reduction; e.g. control mitigates hazard.
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
		field.Enum("category").
			Values(knowledgeEntityCategories...).
			Comment("Stable semantic category used by graph queries, generated views, and agent reasoning.").
			Immutable(),
		field.String("kind").NotEmpty().
			Comment("Canonical domain type within the entity category.").
			Immutable(),
	}
}

func (KnowledgeEntity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("aliases", KnowledgeSubjectAlias.Type).
			Ref("entity"),
		edge.From("linking_attributes", KnowledgeEntityLinkingAttribute.Type).
			Ref("entity"),
		edge.From("source_relationships", KnowledgeRelationship.Type).
			Ref("source_entity"),
		edge.From("target_relationships", KnowledgeRelationship.Type).
			Ref("target_entity"),
	}
}

func (KnowledgeEntity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "category", "kind"),
	}
}

type KnowledgeEntityLinkingAttribute struct {
	ent.Schema
}

func (KnowledgeEntityLinkingAttribute) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (KnowledgeEntityLinkingAttribute) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("entity_id", uuid.UUID{}).Immutable(),
		field.String("attribute").NotEmpty().Immutable(),
		field.String("value").NotEmpty().Immutable(),
	}
}

func (KnowledgeEntityLinkingAttribute) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entity", KnowledgeEntity.Type).
			Required().Unique().Immutable().Field("entity_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (KnowledgeEntityLinkingAttribute) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "attribute", "value").Unique(),
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
		field.Enum("predicate").
			Values(knowledgeRelationshipPredicates...).
			Comment("Canonical directional meaning from source entity to target entity.").
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
		index.Fields("tenant_id", "predicate", "source_entity_id", "target_entity_id").Unique(),
		index.Fields("tenant_id", "source_entity_id"),
		index.Fields("tenant_id", "target_entity_id"),
		index.Fields("tenant_id", "predicate"),
	}
}

type KnowledgeSubjectAlias struct {
	ent.Schema
}

func (KnowledgeSubjectAlias) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Checks: map[string]string{
			"knowledge_subject_alias_exactly_one_subject": "(subject_kind = 'entity' AND entity_id IS NOT NULL AND relationship_id IS NULL) OR (subject_kind = 'relationship' AND relationship_id IS NOT NULL AND entity_id IS NULL)",
		}},
	}
}

func (KnowledgeSubjectAlias) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		ProviderResourceReferenceMixin{},
	}
}

func (KnowledgeSubjectAlias) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),

		field.Enum("subject_kind").
			Values("entity", "relationship").
			Immutable(),

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
		index.Fields("tenant_id", "provider", "provider_namespace", "provider_resource_ref").Unique(),
		index.Fields("tenant_id", "entity_id"),
		index.Fields("tenant_id", "relationship_id"),
	}
}

func (KnowledgeSubjectAlias) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entity", KnowledgeEntity.Type).
			Unique().
			Field("entity_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("relationship", KnowledgeRelationship.Type).
			Unique().
			Field("relationship_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),

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
