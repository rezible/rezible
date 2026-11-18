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

func (Situation) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Checks: map[string]string{
			"situation_merged_into_exactly_when_merged": "(close_reason IS NOT DISTINCT FROM 'merged') = (merged_into_id IS NOT NULL)",
		}},
	}
}

func (Situation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (Situation) Fields() []ent.Field {
	return []ent.Field{
		field.Int("number").Positive().Immutable(),
		field.String("title").NotEmpty(),
		field.Text("summary").Optional(),
		field.Time("opened_at").
			Comment("The earliest start of its observations; lowered when an earlier one joins, never raised"),
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
		field.UUID("investigation_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("The investigation of its observations; several situations may share one"),
		field.UUID("merged_into_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("The situation it was merged into, which is not itself merged"),
	}
}

func (Situation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("investigation", Investigation.Type).
			Field("investigation_id").
			Unique(),
		edge.From("observation_groups", SituationObservationGroup.Type).
			Ref("situation"),
		edge.From("incidents", Incident.Type).
			Ref("situations"),
		edge.From("observations", SituationObservation.Type).
			Ref("situation"),
		edge.From("actions", SituationAction.Type).
			Ref("situation"),
		edge.From("judgments", SituationJudgment.Type).
			Ref("situation"),
		edge.To("notices", SituationNotice.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		// A target with merged sources cannot be deleted: setting them null would break the merge check.
		edge.To("merged_sources", Situation.Type).
			Annotations(entsql.OnDelete(entsql.NoAction)).
			From("merged_into").
			Unique().
			Field("merged_into_id"),
	}
}

func (Situation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "number").
			Unique().
			StorageKey("situations_number_key"),
		index.Fields("tenant_id", "opened_at"),
		index.Fields("tenant_id", "closed_at"),
		index.Fields("tenant_id", "investigation_id"),
		index.Fields("tenant_id", "merged_into_id"),
	}
}

// SituationObservationGroup is a titled group of some of a situation's observations.
type SituationObservationGroup struct {
	ent.Schema
}

func (SituationObservationGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SituationObservationGroup) Fields() []ent.Field {
	return []ent.Field{
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
		edge.From("observations", SituationObservation.Type).
			Ref("observation_group"),
	}
}

// SituationObservation links an observation, such as an alert episode, to a situation. An observation can belong to
// several situations, closed ones included, but to each only once.
type SituationObservation struct {
	ent.Schema
}

func (SituationObservation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
	}
}

func (SituationObservation) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Checks: map[string]string{
			"situation_observation_observed_revision_nonnegative": "observed_revision >= 0",
			// Each observation kind's reference is one nullable column; exactly one is set.
			"situation_observation_exactly_one_reference": "num_nonnulls(alert_episode_id) = 1",
		}},
	}
}

func (SituationObservation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("situation_id", uuid.UUID{}),
		field.UUID("observation_group_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("A group of the same situation"),
		field.UUID("alert_episode_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("attached_at").
			Comment("Processing time of attachment"),
		field.Text("match_explanation").
			Optional().
			Comment("A person's reason for a manual attachment or merge"),
		field.Int("observed_revision").
			NonNegative().
			Default(0).
			Comment("The evidence revision the situation's investigation last saw"),
	}
}

func (SituationObservation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("situation", Situation.Type).
			Field("situation_id").
			Required().
			Unique(),
		edge.To("observation_group", SituationObservationGroup.Type).
			Field("observation_group_id").
			Unique(),
		edge.From("alert_episode", AlertEpisode.Type).
			Ref("situation_observations").
			Field("alert_episode_id").
			Unique(),
	}
}

func (SituationObservation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "situation_id", "alert_episode_id").Unique(),
		index.Fields("tenant_id", "alert_episode_id"),
	}
}

// SituationAction is an append-only log of lifecycle decisions, by people or by the system. Creation,
// attachment and evidence changes are not actions.
type SituationAction struct {
	ent.Schema
}

func (SituationAction) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
	}
}

func (SituationAction) Fields() []ent.Field {
	return []ent.Field{
		// Time-ordered, so actions recorded at one processing time keep their order.
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
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
	}
}

func (SituationJudgment) Fields() []ent.Field {
	return []ent.Field{
		// Time-ordered, so judgments recorded at one processing time keep their order.
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

// SituationSignalAttention is how much a signal source may raise situations. A source references it through
// its own edge; a source without one has the default level.
type SituationSignalAttention struct {
	ent.Schema
}

func (SituationSignalAttention) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
	}
}

func (SituationSignalAttention) Fields() []ent.Field {
	return []ent.Field{
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
