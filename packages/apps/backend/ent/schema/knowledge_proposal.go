package schema

import (
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
		field.String("name_a").NotEmpty().Immutable().
			Comment("The value that sorts first."),
		field.String("name_b").NotEmpty().Immutable(),
		field.Enum("status").
			Values("pending", "confirmed", "rejected").
			Default("pending"),
		field.Enum("origin").
			Values("name_similarity", "identity_conflict").
			Immutable(),
		field.String("rationale").NotEmpty().MaxLen(1000).Immutable().
			Comment("Rezible's own sentence saying why the names may match."),
		field.UUID("evidence_id", uuid.UUID{}).Immutable().
			Comment("Evidence that brought the name which caused the proposal."),
	}
}

func (KnowledgeProposal) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("evidence", KnowledgeEvidence.Type).
			Required().
			Unique().
			Immutable().
			Field("evidence_id"),
	}
}

func (KnowledgeProposal) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "attribute", "name_a", "name_b").Unique(),
		index.Fields("tenant_id", "status"),
	}
}
