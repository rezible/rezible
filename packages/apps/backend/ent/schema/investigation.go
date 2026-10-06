package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
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
		field.UUID("agent_session_id", uuid.UUID{}).Immutable(),
		field.UUID("system_analysis_id", uuid.UUID{}).Immutable(),
	}
}

func (Investigation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_session_id").Unique(),
		index.Fields("tenant_id", "system_analysis_id").Unique(),
	}
}

func (Investigation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("agent_session", AgentSession.Type).Ref("investigation").
			Unique().
			Required().
			Immutable().
			Field("agent_session_id"),
		edge.From("system_analysis", SystemAnalysis.Type).Ref("investigation").
			Unique().
			Required().
			Immutable().
			Field("system_analysis_id"),

		edge.To("situations", SituationInvestigation.Type),

		edge.To("user_inputs", InvestigationUserInput.Type),
		edge.To("evidence_revisions", InvestigationEvidenceRevision.Type),

		edge.To("hypotheses", InvestigationHypothesis.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("findings", InvestigationFinding.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("reports", InvestigationReport.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

type InvestigationUserInput struct {
	ent.Schema
}

func (InvestigationUserInput) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		CreatedAtMixin{},
	}
}

func (InvestigationUserInput) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("investigation_id", uuid.UUID{}).Immutable(),
		field.UUID("user_id", uuid.UUID{}).Immutable(),
		field.Text("text").NotEmpty().Immutable(),
		field.String("key").NotEmpty().Immutable(),
		field.UUID("agent_turn_id", uuid.UUID{}).Optional().Nillable(),
	}
}

func (InvestigationUserInput) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("investigation", Investigation.Type).
			Ref("user_inputs").
			Unique().
			Required().
			Immutable().
			Field("investigation_id"),
		edge.To("agent_turn", AgentTurn.Type).
			Unique().
			Field("agent_turn_id").
			Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.From("findings", InvestigationFinding.Type).
			Ref("user_input"),
	}
}

func (InvestigationUserInput) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("investigation_id", "key").Unique(),
		index.Fields("investigation_id", "agent_turn_id").Unique(),
		index.Fields("investigation_id", "created_at", "id"),
	}
}

type InvestigationEvidenceRevision struct {
	ent.Schema
}

func (InvestigationEvidenceRevision) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		CreatedAtMixin{},
	}
}

func (InvestigationEvidenceRevision) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("investigation_id", uuid.UUID{}).Immutable(),
		field.Text("explanation").NotEmpty().Immutable(),
		field.String("key").NotEmpty().Immutable(),
		field.UUID("agent_turn_id", uuid.UUID{}).Optional().Nillable(),
	}
}

func (InvestigationEvidenceRevision) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("investigation_id", "key").Unique(),
		index.Fields("investigation_id", "agent_turn_id").Unique(),
		index.Fields("investigation_id", "created_at", "id"),
	}
}

func (InvestigationEvidenceRevision) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("investigation", Investigation.Type).
			Ref("evidence_revisions").
			Unique().
			Required().
			Immutable().
			Field("investigation_id"),
		edge.To("agent_turn", AgentTurn.Type).
			Unique().
			Field("agent_turn_id").
			Annotations(entsql.OnDelete(entsql.Restrict)),
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
		field.String("key").NotEmpty().Immutable(),
	}
}

func (InvestigationHypothesis) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "investigation_id", "key").Unique(),
	}
}

func (InvestigationHypothesis) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("investigation", Investigation.Type).
			Ref("hypotheses").
			Unique().
			Required().
			Immutable().
			Field("investigation_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("versions", InvestigationHypothesisVersion.Type).
			Ref("hypothesis"),
	}
}

type InvestigationHypothesisVersion struct {
	ent.Schema
}

func (InvestigationHypothesisVersion) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		CreatedAtMixin{},
		FingerprintMixin{},
	}
}

func (InvestigationHypothesisVersion) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("hypothesis_id", uuid.UUID{}).Immutable(),
		field.UUID("agent_turn_id", uuid.UUID{}).Immutable(),
		field.String("title").NotEmpty().Immutable(),
		field.Enum("status").
			Values("open", "supported", "disproven", "inconclusive").
			Immutable(),
		field.Text("justification").NotEmpty().Immutable(),
	}
}

func (InvestigationHypothesisVersion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("hypothesis", InvestigationHypothesis.Type).
			Required().
			Unique().
			Immutable().
			Field("hypothesis_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("agent_turn", AgentTurn.Type).
			Required().
			Unique().
			Immutable().
			Field("agent_turn_id"),
		edge.From("output_references", InvestigationOutputReference.Type).
			Ref("hypothesis_version"),
	}
}

func (InvestigationHypothesisVersion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_turn_id", "fingerprint").Unique(),
		index.Fields("tenant_id", "hypothesis_id", "agent_turn_id", "created_at", "id"),
	}
}

type InvestigationOutputReference struct {
	ent.Schema
}

func (InvestigationOutputReference) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}, TenantMixin{}}
}

func (InvestigationOutputReference) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Checks: map[string]string{
		"investigation_output_reference_exactly_one_owner": "num_nonnulls(report_id, finding_version_id, hypothesis_version_id) = 1",
	}}}
}

func (InvestigationOutputReference) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("report_id", uuid.UUID{}).Optional().Nillable().Immutable(),
		field.UUID("finding_version_id", uuid.UUID{}).Optional().Nillable().Immutable(),
		field.UUID("hypothesis_version_id", uuid.UUID{}).Optional().Nillable().Immutable(),
		field.UUID("knowledge_evidence_id", uuid.UUID{}).Immutable(),
	}
}

func (InvestigationOutputReference) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "report_id"),
		index.Fields("tenant_id", "finding_version_id"),
		index.Fields("tenant_id", "hypothesis_version_id"),
	}
}

func (InvestigationOutputReference) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("report", InvestigationReport.Type).
			Unique().
			Immutable().
			Field("report_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("finding_version", InvestigationFindingVersion.Type).
			Unique().
			Immutable().
			Field("finding_version_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("hypothesis_version", InvestigationHypothesisVersion.Type).
			Unique().
			Immutable().
			Field("hypothesis_version_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("knowledge_evidence", KnowledgeEvidence.Type).
			Unique().
			Required().
			Immutable().
			Field("knowledge_evidence_id").
			Annotations(entsql.OnDelete(entsql.NoAction)),
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
		field.String("key").NotEmpty().Immutable(),
		field.UUID("user_input_id", uuid.UUID{}).Optional().Nillable().Immutable(),
	}
}

func (InvestigationFinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "investigation_id", "key").Unique(),
		index.Fields("tenant_id", "user_input_id").Unique(),
	}
}

func (InvestigationFinding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("investigation", Investigation.Type).
			Ref("findings").
			Unique().
			Required().
			Immutable().
			Field("investigation_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("user_input", InvestigationUserInput.Type).
			Unique().
			Immutable().
			Field("user_input_id").
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.From("versions", InvestigationFindingVersion.Type).
			Ref("finding"),
	}
}

type InvestigationFindingVersion struct {
	ent.Schema
}

func (InvestigationFindingVersion) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		CreatedAtMixin{},
		FingerprintMixin{},
	}
}

func (InvestigationFindingVersion) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("finding_id", uuid.UUID{}).Immutable(),
		field.UUID("agent_turn_id", uuid.UUID{}).Immutable(),
		field.String("title").NotEmpty().Immutable(),
		field.Text("body").NotEmpty().Immutable(),
	}
}

func (InvestigationFindingVersion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("finding", InvestigationFinding.Type).
			Required().
			Unique().
			Immutable().
			Field("finding_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("agent_turn", AgentTurn.Type).
			Required().
			Unique().
			Immutable().
			Field("agent_turn_id"),
		edge.From("output_references", InvestigationOutputReference.Type).Ref("finding_version"),
		edge.From("outgoing_links", InvestigationFindingVersionLink.Type).Ref("source_version"),
		edge.From("incoming_links", InvestigationFindingVersionLink.Type).Ref("target_version"),
	}
}

func (InvestigationFindingVersion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_turn_id", "fingerprint").Unique(),
		index.Fields("tenant_id", "finding_id", "agent_turn_id", "created_at", "id"),
	}
}

type InvestigationFindingVersionLink struct {
	ent.Schema
}

func (InvestigationFindingVersionLink) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}, TenantMixin{}}
}

func (InvestigationFindingVersionLink) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Checks: map[string]string{
		"investigation_finding_version_link_distinct_versions": "source_version_id <> target_version_id",
	}}}
}

func (InvestigationFindingVersionLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("source_version_id", uuid.UUID{}).Immutable(),
		field.Enum("relation").
			Values("supports", "contradicts", "invalidates").
			Immutable(),
		field.UUID("target_version_id", uuid.UUID{}).Immutable(),
	}
}

func (InvestigationFindingVersionLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("source_version", InvestigationFindingVersion.Type).
			Required().
			Unique().
			Immutable().
			Field("source_version_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("target_version", InvestigationFindingVersion.Type).
			Required().
			Unique().
			Immutable().
			Field("target_version_id"),
	}
}

func (InvestigationFindingVersionLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "source_version_id", "target_version_id", "relation").Unique(),
		index.Fields("tenant_id", "target_version_id"),
	}
}

type InvestigationReport struct {
	ent.Schema
}

func (InvestigationReport) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		CreatedAtMixin{},
		FingerprintMixin{},
	}
}

func (InvestigationReport) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("investigation_id", uuid.UUID{}).Immutable(),
		field.UUID("agent_turn_id", uuid.UUID{}).Immutable(),
		field.Text("text").NotEmpty().Immutable(),
		field.Text("summary").Default("").Immutable(),
	}
}

func (InvestigationReport) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_turn_id", "fingerprint").Unique(),
		index.Fields("tenant_id", "investigation_id", "agent_turn_id", "created_at", "id"),
	}
}

func (InvestigationReport) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("investigation", Investigation.Type).
			Ref("reports").
			Unique().
			Required().
			Immutable().
			Field("investigation_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("agent_turn", AgentTurn.Type).
			Unique().
			Required().
			Immutable().
			Field("agent_turn_id"),
		edge.From("output_references", InvestigationOutputReference.Type).
			Ref("report"),
	}
}
