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

func (Situation) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Checks: map[string]string{
			"situation_open_state_consistent":   "(status = 'open' AND closed_at IS NULL AND close_reason IS NULL) OR status <> 'open'",
			"situation_closed_state_consistent": "(status = 'closed' AND closed_at IS NOT NULL AND close_reason IS NOT NULL) OR status <> 'closed'",
		}},
	}
}

func (Situation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("knowledge_entity_id", uuid.UUID{}).Immutable(),
		field.String("title").NotEmpty(),
		field.Int("evidence_revision").Positive().Default(1),
		field.Text("summary").Optional(),
		field.Enum("status").Values("open", "closed").Default("open"),
		field.Time("opened_at"),
		field.Time("closed_at").Optional().Nillable(),
		field.Enum("close_reason").Values("stabilized", "dismissed").Optional().Nillable(),
	}
}

func (Situation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("knowledge_entity", KnowledgeEntity.Type).
			Unique().
			Required().
			Immutable().
			Field("knowledge_entity_id"),
		edge.To("investigations", SituationInvestigation.Type),
		edge.From("alert_episodes", AlertEpisode.Type).
			Ref("situations").
			Through("alert_episode_links", AlertEpisodeSituation.Type),
		edge.To("hazard_assessments", SituationHazardAssessment.Type),
		edge.From("incidents", Incident.Type).
			Ref("situations"),
	}
}

func (Situation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "knowledge_entity_id").Unique(),
		index.Fields("tenant_id", "status", "opened_at"),
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
		field.UUID("system_analysis_id", uuid.UUID{}).Immutable(),
		field.UUID("agent_session_id", uuid.UUID{}).Immutable(),
		field.Int("completed_revision").NonNegative().Default(0),
		field.Int("requested_revision").NonNegative().Default(0),
		field.UUID("requested_turn_id", uuid.UUID{}).Optional().Nillable(),
		field.JSON("report", &schematypes.SituationInvestigationReport{}).
			SchemaType(schemaTypeJsonB).
			Optional(),
	}
}

func (SituationInvestigation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("requested_turn", AgentTurn.Type).
			Unique().
			Field("requested_turn_id"),
		edge.From("situation", Situation.Type).
			Ref("investigations").
			Unique().
			Required().
			Immutable().
			Field("situation_id"),
		edge.From("system_analysis", SystemAnalysis.Type).
			Ref("situation_investigation").
			Unique().
			Required().
			Immutable().
			Field("system_analysis_id"),
		edge.From("agent_session", AgentSession.Type).
			Ref("situation_investigation").
			Unique().
			Required().
			Immutable().
			Field("agent_session_id"),
	}
}

func (SituationInvestigation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "situation_id"),
		index.Fields("tenant_id", "system_analysis_id").Unique(),
		index.Fields("tenant_id", "agent_session_id").Unique(),
	}
}

// AlertEpisodeSituation is the tenant-scoped association between alert evidence and situations.
type AlertEpisodeSituation struct {
	ent.Schema
}

func (AlertEpisodeSituation) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}, TenantMixin{}}
}

func (AlertEpisodeSituation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("alert_episode_id", uuid.UUID{}),
		field.UUID("situation_id", uuid.UUID{}),
	}
}

func (AlertEpisodeSituation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("alert_episode", AlertEpisode.Type).Required().Unique().Field("alert_episode_id"),
		edge.To("situation", Situation.Type).Required().Unique().Field("situation_id"),
	}
}

func (AlertEpisodeSituation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "alert_episode_id", "situation_id").Unique(),
		index.Fields("tenant_id", "situation_id"),
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
		edge.From("situation", Situation.Type).
			Ref("hazard_assessments").
			Unique().
			Required().
			Immutable().
			Field("situation_id"),
		edge.From("system_hazard", SystemHazard.Type).
			Ref("situation_assessments").
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
