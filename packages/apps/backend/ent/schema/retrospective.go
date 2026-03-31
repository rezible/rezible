package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Retrospective holds the schema definition for the Retrospective entity.
type Retrospective struct {
	ent.Schema
}

func (Retrospective) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AccessScopeMixin{},
		TenantMixin{},
	}
}

// Fields of the Retrospective.
func (Retrospective) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.UUID("document_id", uuid.UUID{}),
		field.UUID("system_analysis_id", uuid.UUID{}).Optional(),
		field.Enum("kind").Values("simple", "full"),
		field.Enum("state").Values("draft", "in_review", "meeting", "closed"),
	}
}

// Edges of the Retrospective.
func (Retrospective) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incident", Incident.Type).
			Ref("retrospective").
			Field("incident_id").
			Unique().
			Required(),

		edge.From("document", Document.Type).
			Ref("retrospective").
			Field("document_id").
			Unique().
			Required(),
		edge.From("comments", RetrospectiveComment.Type).
			Ref("retrospective"),
		edge.From("system_analysis", SystemAnalysis.Type).
			Ref("retrospective").Unique().Field("system_analysis_id"),
	}
}

type RetrospectiveComment struct {
	ent.Schema
}

func (RetrospectiveComment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AccessScopeMixin{},
		TenantMixin{},
	}
}

func (RetrospectiveComment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("retrospective_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("retrospective_review_id", uuid.UUID{}).Optional(),
		field.UUID("parent_reply_id", uuid.UUID{}).Optional(),
		field.Bytes("content"),
	}
}

func (RetrospectiveComment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("retrospective", Retrospective.Type).
			Field("retrospective_id").
			Required().
			Unique(),
		edge.To("user", User.Type).
			Field("user_id").
			Required().
			Unique(),
		edge.To("review", RetrospectiveReview.Type).
			Field("retrospective_review_id").
			Unique(),
		edge.To("replies", RetrospectiveComment.Type).
			From("parent").
			Field("parent_reply_id").
			Unique(),
	}
}

// RetrospectiveReview holds the schema definition for the RetrospectiveReview entity.
type RetrospectiveReview struct {
	ent.Schema
}

func (RetrospectiveReview) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AccessScopeMixin{},
		TenantMixin{},
	}
}

// Fields of the RetrospectiveReview.
func (RetrospectiveReview) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("retrospective_id", uuid.UUID{}),
		field.UUID("comment_id", uuid.UUID{}),
		field.UUID("requester_id", uuid.UUID{}),
		field.UUID("reviewer_id", uuid.UUID{}),
		field.Enum("state").Values("waiting", "request_changes", "approved"),
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
		edge.To("comment", RetrospectiveComment.Type).
			Field("comment_id").
			Unique().
			Required(),
	}
}
