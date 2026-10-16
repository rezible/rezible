package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// KnowledgeProposal is a match Rezible suggests and people decide on: two entities of one category and kind,
// named by two values of a linking attribute, are the same thing. It names values, not entities, so it reads
// the same after the entities are merged.
type KnowledgeProposal struct {
	ent.Schema
}

func (KnowledgeProposal) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (KnowledgeProposal) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Enum("category").
			Values(knowledgeEntityCategories...).
			Immutable(),
		field.String("kind").NotEmpty().Immutable(),
		field.String("attribute").NotEmpty().Immutable().
			Comment("Linking attribute whose two values name the matched entities, such as service.name."),
		field.String("value_a").NotEmpty().Immutable().
			Comment("The value that sorts first."),
		field.String("value_b").NotEmpty().Immutable(),
		field.Enum("status").
			Values("pending", "confirmed", "rejected").
			Default("pending"),
		field.Enum("origin").
			Values("name_similarity", "identity_conflict").
			Immutable(),
		field.String("rationale").NotEmpty().MaxLen(1000).Immutable().
			Comment("Rezible's own sentence saying why the two values may name one entity."),
		field.UUID("evidence_id", uuid.UUID{}).Immutable().
			Comment("Evidence that brought the value which caused the proposal."),
	}
}

func (KnowledgeProposal) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("evidence", KnowledgeEvidence.Type).
			Required().
			Unique().
			Immutable().
			Field("evidence_id"),
		edge.From("decisions", KnowledgeProposalDecision.Type).
			Ref("proposal"),
	}
}

func (KnowledgeProposal) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "attribute", "value_a", "value_b").Unique(),
		index.Fields("tenant_id", "status"),
	}
}

// KnowledgeProposalDecision is one person's decision on a proposal. Decisions are kept as its history.
type KnowledgeProposalDecision struct {
	ent.Schema
}

func (KnowledgeProposalDecision) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (KnowledgeProposalDecision) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("proposal_id", uuid.UUID{}).Immutable(),
		field.Enum("status").
			Values("confirmed", "rejected").
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).Immutable(),
		field.UUID("implied_by_proposal_id", uuid.UUID{}).
			Optional().
			Nillable().
			Immutable().
			Comment("The proposal whose confirmation merged this one's values onto one entity, when the decision follows from it."),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (KnowledgeProposalDecision) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("proposal", KnowledgeProposal.Type).
			Required().
			Unique().
			Immutable().
			Field("proposal_id"),
		edge.To("user", User.Type).
			Required().
			Unique().
			Immutable().
			Field("user_id"),
		edge.To("implied_by_proposal", KnowledgeProposal.Type).
			Unique().
			Immutable().
			Field("implied_by_proposal_id"),
	}
}

func (KnowledgeProposalDecision) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "proposal_id"),
	}
}
