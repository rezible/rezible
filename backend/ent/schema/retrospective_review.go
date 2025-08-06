package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// RetrospectiveReview holds the schema definition for the RetrospectiveReview entity.
type RetrospectiveReview struct {
	ent.Schema
}

func (RetrospectiveReview) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

var (
	retroReviewStates = []string{"waiting", "request_changes", "approved"}
)

// Fields of the RetrospectiveReview.
func (RetrospectiveReview) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("retrospective_id", uuid.UUID{}),
		field.UUID("requester_id", uuid.UUID{}),
		field.UUID("reviewer_id", uuid.UUID{}),
		field.Enum("state").Values(retroReviewStates...),
	}
}

// Edges of the RetrospectiveReview.
func (RetrospectiveReview) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("retrospective", Retrospective.Type).
			Field("retrospective_id").
			Required().
			Unique(),
		edge.To("requester", User.Type).
			Field("requester_id").
			Required().
			Unique(),
		edge.To("reviewer", User.Type).
			Field("reviewer_id").
			Required().
			Unique(),
		edge.To("discussion", RetrospectiveDiscussion.Type).
			Unique(),
	}
}
