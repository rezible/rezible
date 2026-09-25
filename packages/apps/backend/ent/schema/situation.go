package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
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
		field.Time("opened_at"),
		field.Time("closed_at").Optional().Nillable(),
		field.Enum("close_reason").Values("stabilized", "dismissed").Optional().Nillable(),
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
	}
}

func (Situation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "opened_at"),
	}
}

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
		edge.To("events", NormalizedEvent.Type),
		edge.To("alert_episodes", AlertEpisode.Type),
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
