package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type SystemHazard struct {
	ent.Schema
}

func (SystemHazard) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SystemHazard) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("title").NotEmpty(),
		field.Text("description").Optional(),
		field.Text("potential_consequences").Optional(),
		field.Enum("status").Values("active", "retired").Default("active"),
	}
}

func (SystemHazard) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("risk_assessments", SystemHazardRiskAssessment.Type),
		edge.To("situation_assessments", SituationHazardAssessment.Type),
	}
}

func (SystemHazard) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "status", "title"),
	}
}

type SystemHazardRiskAssessment struct {
	ent.Schema
}

func (SystemHazardRiskAssessment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SystemHazardRiskAssessment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("system_hazard_id", uuid.UUID{}).Immutable(),
		field.Int("revision").Positive().Immutable(),
		field.String("likelihood").NotEmpty().Immutable(),
		field.String("consequence").NotEmpty().Immutable(),
		field.String("risk_level").NotEmpty().Immutable(),
		field.Text("rationale").Optional().Immutable(),
		field.Time("assessed_at").Immutable(),
	}
}

func (SystemHazardRiskAssessment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("system_hazard", SystemHazard.Type).
			Ref("risk_assessments").
			Unique().
			Required().
			Immutable().
			Field("system_hazard_id"),
	}
}

func (SystemHazardRiskAssessment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "system_hazard_id", "revision").Unique(),
		index.Fields("tenant_id", "system_hazard_id", "assessed_at"),
	}
}
