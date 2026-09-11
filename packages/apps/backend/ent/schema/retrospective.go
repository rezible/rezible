package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Retrospective holds the schema definition for the Retrospective entity.
type Retrospective struct {
	ent.Schema
}

func (Retrospective) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

// Fields of the Retrospective.
func (Retrospective) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.UUID("document_id", uuid.UUID{}),
		field.UUID("system_analysis_id", uuid.UUID{}),
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
		edge.From("discussion_threads", DiscussionThread.Type).
			Ref("retrospective"),
		edge.From("reviews", Review.Type).Ref("retrospective"),
		edge.To("system_analysis", SystemAnalysis.Type).
			Field("system_analysis_id").
			Unique().
			Required(),
	}
}

type DiscussionThread struct {
	ent.Schema
}

func (DiscussionThread) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (DiscussionThread) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("analysis_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("retrospective_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("kind").Values("comment", "question"),
		field.Enum("target_kind").Values("finding", "knowledge_entity", "knowledge_relationship", "normalized_event").Optional().Nillable(),
		field.UUID("target_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("resolution_state").Values("open", "resolved").Optional().Nillable(),
		field.UUID("resolved_by_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("resolved_at").Optional().Nillable(),
		field.Text("resolution_note").Optional().Nillable(),
	}
}

func (DiscussionThread) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{
			Checks: map[string]string{
				"discussion_thread_exactly_one_owner": "(analysis_id IS NOT NULL AND retrospective_id IS NULL) OR (analysis_id IS NULL AND retrospective_id IS NOT NULL)",
			},
		},
	}
}

func (DiscussionThread) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("analysis", SystemAnalysis.Type).Field("analysis_id").Unique(),
		edge.To("retrospective", Retrospective.Type).Field("retrospective_id").Unique(),
		edge.To("user", User.Type).
			Field("user_id").
			Required().
			Unique(),
		edge.From("comments", DiscussionComment.Type).Ref("thread"),
	}
}

type DiscussionComment struct {
	ent.Schema
}

func (DiscussionComment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (DiscussionComment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("thread_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Text("content"),
		field.UUID("parent_id", uuid.UUID{}).Optional().Nillable(),
	}
}

func (DiscussionComment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("thread", DiscussionThread.Type).Field("thread_id").Required().Unique(),
		edge.To("user", User.Type).Field("user_id").Required().Unique(),
		edge.To("parent", DiscussionComment.Type).Field("parent_id").Unique(),
		edge.From("replies", DiscussionComment.Type).Ref("parent"),
		edge.From("reviews", Review.Type).Ref("comment"),
	}
}

// Review holds review metadata for a retrospective or system-analysis entry.
type Review struct {
	ent.Schema
}

func (Review) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (Review) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{
			Checks: map[string]string{
				"review_exactly_one_subject": "num_nonnulls(retrospective_id, analysis_entry_id) = 1",
			},
		},
	}
}

func (Review) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("retrospective_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("analysis_entry_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("comment_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("requester_id", uuid.UUID{}),
		field.UUID("reviewer_id", uuid.UUID{}),
		field.Enum("state").Values("waiting", "request_changes", "approved"),
	}
}

func (Review) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("retrospective", Retrospective.Type).
			Field("retrospective_id").
			Unique(),
		edge.To("analysis_entry", SystemAnalysisEntry.Type).
			Field("analysis_entry_id").
			Unique(),
		edge.To("requester", User.Type).
			Field("requester_id").
			Required().
			Unique(),
		edge.To("reviewer", User.Type).
			Field("reviewer_id").
			Required().
			Unique(),
		edge.To("comment", DiscussionComment.Type).
			Field("comment_id").
			Unique(),
	}
}
