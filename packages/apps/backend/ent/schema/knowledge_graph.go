package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/schema/schematypes"
)

var (
	knowledgeEntityCategories = []string{
		"system_function", // What the business needs done, whatever implements it; e.g. capability.
		"system",          // A coherent software or business system, including external ones; e.g. subsystem, external_system.
		"container",       // An independently deployable or runnable unit; e.g. service, worker, database, cache, message_queue, web_app, gateway.
		"component",       // A part inside a container; e.g. module, handler, scheduled_job.
		"infrastructure",  // Where things run or are deployed; e.g. environment, cloud_account, region, kubernetes_cluster, kubernetes_namespace, load_balancer.
		"code",            // A source artifact; e.g. repository, package, source_file.
		"actor",           // A person or a group of people; e.g. user, team, department, customer_segment, vendor.
		"process",         // A business or operational workflow, or one of its steps; e.g. business_process, operational_process, step.
		"concern",         // A force on the system: risk, hazard, requirement, regulation, target; e.g. system_hazard, risk, regulation, policy, slo.
		"decision",        // A chosen design or operating trade-off; e.g. adr, mitigation_choice.
		"signal",          // A definition that watches something; e.g. alert, synthetic_check, dashboard.
		"event",           // A time-bound occurrence; e.g. deployment, code_change, incident.
	}

	// Each predicate's family, dependency, roll-up and allowed endpoints live in pkg/knowledgegraph/model.go.
	knowledgeRelationshipPredicates = []string{
		"contains",        // The target is part of the source's structure; e.g. Commerce Core contains checkout.
		"runs_on",         // The source executes or is deployed there; e.g. checkout runs_on prod-use1/commerce.
		"built_from",      // The code produces the source; e.g. checkout built_from checkout-service.
		"owns",            // The actor is accountable for it; e.g. Commerce Team owns checkout.
		"member_of",       // A person belongs to a group (several allowed); e.g. Ava member_of Search Platform.
		"participates_in", // The actor takes part in the workflow; e.g. Commerce Team participates_in Checkout flow.
		"uses",            // People rely on it directly; e.g. Customers uses Customer Web App.
		"supports",        // The workflow needs it; e.g. checkout supports Checkout flow.
		"calls",           // Request and response; e.g. web_app calls public_api_gateway.
		"reads_from",      // The source reads its data; e.g. search_api reads_from elasticsearch_catalog.
		"writes_to",       // The source writes its data; e.g. orders_service writes_to orders_postgres.
		"publishes_to",    // The source sends messages to it; e.g. orders_service publishes_to order_events_queue.
		"consumes_from",   // The source takes messages from it; e.g. fulfilment_worker consumes_from order_events_queue.
		"depends_on",      // The source needs it, and no flow predicate above fits; e.g. checkout depends_on feature_flags.
		"observes",        // The monitor watches it; e.g. "Checkout error rate" observes checkout.
		"constrains",      // A rule or choice limits it; e.g. "Checkout p99 under 800 ms" constrains checkout.
		"addresses",       // The decision responds to the concern; e.g. ADR-12 addresses Duplicate payment capture.
		"mitigates",       // The source reduces the risk; e.g. payments_service mitigates Duplicate payment capture.
		"impacts",         // The occurrence affected it; e.g. deploy impacts checkout.
		"touches",         // The change modified it; e.g. code change touches checkout-service.
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
		KnowledgeSubjectStateMixin{},
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
		field.UUID("parent_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Source of the one current contains relationship chosen as the entity's parent; every reader of the tree uses it. Kept current on every write that can change the choice."),
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
		edge.To("children", KnowledgeEntity.Type).
			From("parent").
			Unique().
			Field("parent_id").
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (KnowledgeEntity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "category", "kind"),
		index.Fields("tenant_id", "kind", "state_effective_at"),
		index.Fields("tenant_id", "withdrawn_at"),
		index.Fields("tenant_id", "parent_id"),
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
		field.UUID("entity_id", uuid.UUID{}).
			Comment("Changed only when a confirmed match merges two entities."),
		field.String("attribute").NotEmpty().Immutable(),
		field.String("value").NotEmpty().Immutable(),
	}
}

func (KnowledgeEntityLinkingAttribute) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entity", KnowledgeEntity.Type).
			Required().Unique().Field("entity_id").
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
		KnowledgeSubjectStateMixin{},
	}
}

func (KnowledgeRelationship) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Enum("predicate").
			Values(knowledgeRelationshipPredicates...).
			Comment("Canonical directional meaning from source entity to target entity.").
			Immutable(),
		field.UUID("source_entity_id", uuid.UUID{}).
			Comment("Changed only when a confirmed match merges two entities."),
		field.UUID("target_entity_id", uuid.UUID{}).
			Comment("Changed only when a confirmed match merges two entities."),
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
		index.Fields("tenant_id", "predicate", "source_entity_id", "target_entity_id").Unique(),
		index.Fields("tenant_id", "source_entity_id"),
		index.Fields("tenant_id", "target_entity_id"),
		index.Fields("tenant_id", "predicate"),
		index.Fields("tenant_id", "withdrawn_at"),
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
		field.UUID("merged_by_proposal_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("The confirmed proposal whose merge moved the alias to its current subject."),
		field.Time("created_at").Default(time.Now).Immutable(),

		field.JSON("state", schematypes.KnowledgeAliasState{}).
			Optional().
			SchemaType(schemaTypeJsonB).
			Comment("What this alias's evidence says, field by field, with each field's effective time."),
		field.Time("first_observed_at").Optional().Nillable(),
		field.Time("last_observed_at").Optional().Nillable(),
		field.Time("withdrawn_at").Optional().Nillable().
			Comment("Set when the alias's latest evidence says the subject no longer exists."),
		field.Time("expires_at").Optional().Nillable().
			Comment("When the latest observation stops holding unless observed again; null never expires."),
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

		edge.To("merged_by_proposal", KnowledgeProposal.Type).
			Unique().
			Field("merged_by_proposal_id"),

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
		field.Enum("state_mode").Immutable().
			Values("merge", "replace").
			Default("merge").
			Comment("merge sets the fields stated; replace is a full snapshot that also unsets the properties it omits."),
		field.Time("expires_at").Optional().Nillable().Immutable().
			Comment("Effective time plus the producer's validity: the observation holds until then unless observed again."),
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

// KnowledgeSubjectStateMixin holds a subject's state merged from its aliases, and its observation times.
type KnowledgeSubjectStateMixin struct {
	mixin.Schema
}

func (KnowledgeSubjectStateMixin) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("state", schematypes.KnowledgeGraphSubjectState{}).
			Optional().
			SchemaType(schemaTypeJsonB).
			Comment("Display name, description and properties, merged key by key from the live aliases."),
		field.JSON("state_sources", schematypes.KnowledgeStateSources{}).
			Optional().
			SchemaType(schemaTypeJsonB).
			Comment("For each field of the state, the alias that supplied it and its effective time."),
		field.Time("state_effective_at").
			Optional().
			Nillable().
			Comment("Latest effective time among the state's fields. For event entities this is when the event happened, as currently known: projections set an event's evidence effective time to the event's own time."),
		field.Time("last_observed_at").Optional().Nillable(),
		field.Time("withdrawn_at").Optional().Nillable().
			Comment("Set when the subject has aliases and all are withdrawn. Default reads exclude withdrawn subjects."),
		field.Time("expires_at").Optional().Nillable().
			Comment("When the subject goes stale; null when any live alias never expires."),
	}
}
