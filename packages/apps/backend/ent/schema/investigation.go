package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Investigation struct {
	ent.Schema
}

func (Investigation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (Investigation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("system_analysis_id", uuid.UUID{}).Immutable(),
		field.UUID("agent_session_id", uuid.UUID{}).Immutable(),
	}
}

func (Investigation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("system_analysis", SystemAnalysis.Type).Ref("investigation").
			Unique().
			Required().
			Immutable().
			Field("system_analysis_id"),
		edge.From("agent_session", AgentSession.Type).Ref("investigation").
			Unique().
			Required().
			Immutable().
			Field("agent_session_id"),
		edge.To("situations", SituationInvestigation.Type),
		edge.To("hypotheses", InvestigationHypothesis.Type),
		edge.To("findings", InvestigationFinding.Type),
		edge.To("report", InvestigationReport.Type).
			Unique(),
	}
}

func (Investigation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "system_analysis_id").Unique(),
		index.Fields("tenant_id", "agent_session_id").Unique(),
	}
}

type InvestigationHypothesis struct {
	ent.Schema
}

func (InvestigationHypothesis) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (InvestigationHypothesis) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("investigation_id", uuid.UUID{}).Immutable(),
		field.String("title").NotEmpty(),
		field.Text("body").Optional(),
		field.Text("verdict").Optional().Nillable(),
	}
}

func (InvestigationHypothesis) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("investigation", Investigation.Type).
			Ref("hypotheses").
			Unique().
			Required().
			Immutable().
			Field("investigation_id"),
	}
}

func (InvestigationHypothesis) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "investigation_id"),
	}
}

type InvestigationFinding struct {
	ent.Schema
}

func (InvestigationFinding) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (InvestigationFinding) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("investigation_id", uuid.UUID{}).Immutable(),
		field.String("title").NotEmpty(),
		field.Text("body").Optional(),
	}
}

func (InvestigationFinding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("investigation", Investigation.Type).
			Ref("findings").
			Unique().
			Required().
			Immutable().
			Field("investigation_id"),
	}
}

func (InvestigationFinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "investigation_id"),
	}
}

type InvestigationReport struct {
	ent.Schema
}

func (InvestigationReport) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (InvestigationReport) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("investigation_id", uuid.UUID{}).Immutable(),
		field.UUID("agent_turn_id", uuid.UUID{}).
			Optional().
			Nillable().
			Immutable(),
		field.Text("text").NotEmpty(),
		field.Text("likely_cause").Optional(),
		field.Text("best_next_step").Optional(),
		field.Strings("limitations").Optional(),
		field.Strings("recommended_actions").Optional(),
		field.Strings("suggested_checks").Optional(),
	}
}

func (InvestigationReport) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("investigation", Investigation.Type).
			Ref("report").
			Unique().
			Required().
			Immutable().
			Field("investigation_id"),
		edge.To("agent_turn", AgentTurn.Type).
			Unique().
			Immutable().
			Field("agent_turn_id"),
	}
}

func (InvestigationReport) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_turn_id"),
		index.Fields("tenant_id", "investigation_id", "created_at"),
	}
}
