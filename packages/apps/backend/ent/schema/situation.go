package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent/schema/schematypes"
)

type Situation struct {
	ent.Schema
}

func (Situation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (Situation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("title").NotEmpty(),
		field.Text("summary").Optional(),
		field.Time("opened_at").
			Comment("The earliest start of its signals; lowered when an earlier signal joins, never raised"),
		field.Time("raised_at").
			Optional().
			Nillable().
			Comment("Processing time it was raised; unset while a candidate"),
		field.Time("muted_at").Optional().Nillable(),
		field.Enum("mute_reason").
			Values("not_noteworthy", "expected").
			Optional().
			Nillable(),
		field.Time("hold_until").
			Optional().
			Nillable().
			Comment("Automatic closure is delayed until this time"),
		field.Time("closed_at").Optional().Nillable(),
		field.Enum("close_reason").
			Values("stabilized", "expired", "merged", "dismissed").
			Optional().
			Nillable(),
		field.UUID("seed_entity_id", uuid.UUID{}).
			Immutable().
			Comment("Knowledge entity of the signal that started it; kept after merges"),
		field.UUID("latest_judgment_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Its most recently recorded judgment"),
	}
}

func (Situation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("investigation", SituationInvestigation.Type).
			Unique(),
		edge.From("hazard_assessments", SituationHazardAssessment.Type).
			Ref("situation"),
		edge.From("observation_groups", SituationObservationGroup.Type).
			Ref("situation"),
		edge.From("incidents", Incident.Type).
			Ref("situations"),
		edge.From("signals", SituationSignal.Type).
			Ref("situation"),
		edge.From("entities", SituationEntity.Type).
			Ref("situation"),
		edge.From("links", SituationLink.Type).
			Ref("situation"),
		edge.From("actions", SituationAction.Type).
			Ref("situation"),
		edge.To("latest_judgment", SituationJudgment.Type).
			Field("latest_judgment_id").
			Unique(),
	}
}

func (Situation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "opened_at"),
		index.Fields("tenant_id", "closed_at"),
	}
}

// SituationObservationGroup is a titled group of a situation's signals.
type SituationObservationGroup struct {
	ent.Schema
}

func (SituationObservationGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SituationObservationGroup) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("situation_id", uuid.UUID{}),
		field.String("title").NotEmpty(),
		field.Text("body").Optional(),
	}
}

func (SituationObservationGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("situation", Situation.Type).
			Field("situation_id").
			Required().
			Unique(),
		edge.From("signals", SituationSignal.Type).
			Ref("observation_group"),
	}
}

// SituationSignal is one signal's membership in a situation. A signal is referenced by its knowledge entity
// and belongs to at most one situation, closed ones included.
type SituationSignal struct {
	ent.Schema
}

func (SituationSignal) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SituationSignal) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("situation_id", uuid.UUID{}),
		field.UUID("observation_group_id", uuid.UUID{}).
			Comment("A group of the same situation"),
		field.UUID("knowledge_entity_id", uuid.UUID{}).Immutable(),
		field.String("kind").
			NotEmpty().
			Comment("The signal's kind, copied when attached"),
		field.UUID("source_entity_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Knowledge entity of what produced the signal, copied when attached"),
		field.Time("attached_at").
			Comment("Processing time of attachment"),
		field.Enum("match_kind").
			Values("seed", "shared_entity", "dependency", "dependent", "adjacent", "manual"),
		field.UUID("via_relationship_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("The relationship that justified a one-hop match"),
		field.Text("match_explanation").
			Optional().
			Comment("A person's reason for a manual attachment or merge"),
		field.Int("observed_revision").
			Default(0).
			Comment("The signal revision the situation's investigation last saw"),
	}
}

func (SituationSignal) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("situation", Situation.Type).
			Field("situation_id").
			Required().
			Unique(),
		edge.To("observation_group", SituationObservationGroup.Type).
			Field("observation_group_id").
			Required().
			Unique(),
		edge.To("knowledge_entity", KnowledgeEntity.Type).
			Field("knowledge_entity_id").
			Required().
			Unique().
			Immutable(),
	}
}

func (SituationSignal) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "knowledge_entity_id").Unique(),
		index.Fields("tenant_id", "situation_id"),
		index.Fields("tenant_id", "source_entity_id"),
	}
}

// SituationEntity is one of a situation's runtime entities, recomputed from its signals.
type SituationEntity struct {
	ent.Schema
}

func (SituationEntity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SituationEntity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("situation_id", uuid.UUID{}).Immutable(),
		field.UUID("knowledge_entity_id", uuid.UUID{}).Immutable(),
		field.Bool("matching").
			Comment("Whether a signal that is not broad touches it, so it attracts related signals"),
	}
}

func (SituationEntity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("situation", Situation.Type).
			Field("situation_id").
			Required().
			Unique().
			Immutable(),
		edge.To("knowledge_entity", KnowledgeEntity.Type).
			Field("knowledge_entity_id").
			Required().
			Unique().
			Immutable(),
	}
}

func (SituationEntity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "situation_id", "knowledge_entity_id").Unique(),
		index.Fields("tenant_id", "knowledge_entity_id").
			Annotations(entsql.IndexWhere("matching")),
	}
}

// SituationLink points from a recurrence or a merged situation to the earlier or surviving one.
type SituationLink struct {
	ent.Schema
}

func (SituationLink) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SituationLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("situation_id", uuid.UUID{}).Immutable(),
		field.UUID("linked_situation_id", uuid.UUID{}).Immutable(),
		field.Enum("kind").
			Values("recurrence_of", "merged_into").
			Immutable(),
		field.Time("created_at").Immutable(),
	}
}

func (SituationLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("situation", Situation.Type).
			Field("situation_id").
			Required().
			Unique().
			Immutable(),
		edge.To("linked_situation", Situation.Type).
			Field("linked_situation_id").
			Required().
			Unique().
			Immutable(),
	}
}

func (SituationLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "situation_id", "linked_situation_id", "kind").Unique(),
	}
}

// SituationAction is an append-only log of lifecycle decisions, by people or by the system. Creation,
// attachment and evidence changes are not actions.
type SituationAction struct {
	ent.Schema
}

func (SituationAction) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SituationAction) Fields() []ent.Field {
	return []ent.Field{
		// Time-ordered, so actions recorded at one processing time keep their order.
		field.UUID("id", uuid.UUID{}).Default(func() uuid.UUID {
			return uuid.Must(uuid.NewV7())
		}),
		field.UUID("situation_id", uuid.UUID{}).Immutable(),
		field.Enum("action").
			Values("raised", "muted", "unmuted", "held", "hold_cleared", "closed", "merged").
			Immutable(),
		field.String("reason").Optional().Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Optional().
			Nillable().
			Immutable().
			Comment("Unset for system actions"),
		field.Time("at").Immutable(),
	}
}

func (SituationAction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("situation", Situation.Type).
			Field("situation_id").
			Required().
			Unique().
			Immutable(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Immutable(),
	}
}

func (SituationAction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "situation_id", "at"),
	}
}

// SituationJudgment is one stored evaluation of a candidate: the facts it was judged on, every reason met
// or not, the decision and its explanation. Append-only.
type SituationJudgment struct {
	ent.Schema
}

func (SituationJudgment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SituationJudgment) Fields() []ent.Field {
	return []ent.Field{
		// Time-ordered, so judgments recorded at one processing time keep their order.
		field.UUID("id", uuid.UUID{}).Default(func() uuid.UUID {
			return uuid.Must(uuid.NewV7())
		}),
		field.UUID("situation_id", uuid.UUID{}).Immutable(),
		field.Time("judged_at").
			Immutable().
			Comment("Processing time of the judgment"),
		field.Enum("outcome").
			Values("no_reason", "needs_decision", "raise").
			Immutable(),
		field.Enum("decision").
			Values("hold", "raise").
			Immutable(),
		field.JSON("reasons", []schematypes.SituationReasonResult{}).
			SchemaType(schemaTypeJsonB).
			Immutable().
			Comment("All reasons, met or not"),
		field.JSON("cited_reasons", []schematypes.SituationRaiseReason{}).
			SchemaType(schemaTypeJsonB).
			Immutable().
			Comment("The reasons the decision rests on"),
		field.JSON("facts", schematypes.SituationFacts{}).
			SchemaType(schemaTypeJsonB).
			Immutable(),
		field.Text("explanation").Immutable(),
		field.String("judge").
			Immutable().
			Comment("Which judge decided"),
		field.String("fingerprint").
			Immutable().
			Comment("Hash of the inputs that can change the decision"),
	}
}

func (SituationJudgment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("situation", Situation.Type).
			Field("situation_id").
			Required().
			Unique().
			Immutable(),
	}
}

func (SituationJudgment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "situation_id", "judged_at"),
	}
}

type SituationInvestigation struct {
	ent.Schema
}

func (SituationInvestigation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SituationInvestigation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("situation_id", uuid.UUID{}).Immutable(),
		field.UUID("investigation_id", uuid.UUID{}).Immutable(),
	}
}

func (SituationInvestigation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("situation", Situation.Type).
			Ref("investigation").
			Unique().
			Required().
			Immutable().
			Field("situation_id"),
		edge.From("investigation", Investigation.Type).
			Ref("situations").
			Unique().
			Required().
			Immutable().
			Field("investigation_id"),
	}
}

func (SituationInvestigation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "situation_id").Unique(),
		index.Fields("tenant_id", "investigation_id").Unique(),
	}
}

type SituationHazardAssessment struct {
	ent.Schema
}

func (SituationHazardAssessment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SituationHazardAssessment) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Checks: map[string]string{
			"situation_hazard_assessment_exactly_one_assessor": "num_nonnulls(user_id, agent_turn_id) = 1",
		}},
	}
}

func (SituationHazardAssessment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("situation_id", uuid.UUID{}).Immutable(),
		field.UUID("system_hazard_id", uuid.UUID{}).Immutable(),
		field.Int("revision").Positive().Immutable(),
		field.Enum("status").Values("suspected", "confirmed", "disproven").Immutable(),
		field.Text("summary").NotEmpty().Immutable(),
		field.Time("assessed_at").Immutable(),
		field.UUID("user_id", uuid.UUID{}).Optional().Nillable().Immutable(),
		field.UUID("agent_turn_id", uuid.UUID{}).Optional().Nillable().Immutable(),
	}
}

func (SituationHazardAssessment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("situation", Situation.Type).
			Unique().
			Required().
			Immutable().
			Field("situation_id"),
		edge.To("system_hazard", SystemHazard.Type).
			Unique().
			Required().
			Immutable().
			Field("system_hazard_id"),
		edge.To("user", User.Type).
			Unique().
			Immutable().
			Field("user_id"),
		edge.To("agent_turn", AgentTurn.Type).
			Unique().
			Immutable().
			Field("agent_turn_id"),
	}
}

func (SituationHazardAssessment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "situation_id", "system_hazard_id", "revision").Unique(),
		index.Fields("tenant_id", "situation_id", "assessed_at"),
		index.Fields("tenant_id", "system_hazard_id", "assessed_at"),
	}
}

// SituationSignalAttention is how much a signal source may raise situations. A source references it through
// its own edge; a source without one has the default level.
type SituationSignalAttention struct {
	ent.Schema
}

func (SituationSignalAttention) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (SituationSignalAttention) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.Enum("level").
			Values("default", "watch_only", "join_only").
			Default("default"),
		field.Time("set_at").
			Comment("Processing time the level last changed"),
		field.UUID("set_by_user_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("The person who last changed the level; unset if that user was deleted"),
	}
}

func (SituationSignalAttention) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("set_by_user", User.Type).
			Unique().
			Field("set_by_user_id").
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}
