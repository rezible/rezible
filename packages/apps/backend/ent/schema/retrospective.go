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

var retrospectiveStates = []string{"draft", "in_review", "closed"}

// Fields of the Retrospective.
func (Retrospective) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.UUID("document_id", uuid.UUID{}),
		field.UUID("system_analysis_id", uuid.UUID{}),
		field.Enum("state").Values(retrospectiveStates...),
		field.UUID("owner_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("closed_at").Optional().Nillable(),
		field.UUID("closed_by_id", uuid.UUID{}).Optional().Nillable(),
		field.Text("skip_reason").Optional().Nillable(),
	}
}

// Edges of the Retrospective.
func (Retrospective) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("owner", User.Type).Field("owner_id").Unique(),
		edge.To("closed_by", User.Type).Field("closed_by_id").Unique(),
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
		edge.From("reviews", Review.Type).
			Ref("retrospective"),
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

		field.UUID("target_finding_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("target_entry_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("target_knowledge_entity_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("target_knowledge_relationship_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("target_normalized_event_id", uuid.UUID{}).Optional().Nillable(),

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
				"discussion_thread_at_most_one_target": "num_nonnulls(target_finding_id, target_entry_id, target_knowledge_entity_id, " +
					"target_knowledge_relationship_id, target_normalized_event_id) <= 1",
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

		// A thread keeps its owner and comments when an analysis entry, finding or event it targets is deleted.
		edge.To("target_finding", InvestigationFinding.Type).
			Field("target_finding_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("target_entry", SystemAnalysisEntry.Type).
			Field("target_entry_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("target_normalized_event", NormalizedEvent.Type).
			Field("target_normalized_event_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
		// A knowledge merge re-points these before it deletes the merged entity or folded relationship.
		edge.To("target_knowledge_entity", KnowledgeEntity.Type).
			Field("target_knowledge_entity_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.To("target_knowledge_relationship", KnowledgeRelationship.Type).
			Field("target_knowledge_relationship_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.NoAction)),
	}
}

func (DiscussionThread) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "target_finding_id"),
		index.Fields("tenant_id", "target_entry_id"),
		index.Fields("tenant_id", "target_knowledge_entity_id"),
		index.Fields("tenant_id", "target_knowledge_relationship_id"),
		index.Fields("tenant_id", "target_normalized_event_id"),
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
		field.UUID("parent_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("user_id", uuid.UUID{}),
		field.Text("content"),
	}
}

func (DiscussionComment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("thread", DiscussionThread.Type).
			Required().
			Unique().
			Field("thread_id"),
		edge.To("parent", DiscussionComment.Type).
			Unique().
			Field("parent_id"),
		edge.To("user", User.Type).
			Required().
			Unique().
			Field("user_id"),
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
		field.Text("feedback").Optional().Nillable(),
		field.Time("message_failed_at").Optional().Nillable(),
	}
}

func (Review) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "retrospective_id", "reviewer_id").Unique(),
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
		edge.To("comment", DiscussionComment.Type).
			Field("comment_id").
			Unique(),
		edge.To("requester", User.Type).
			Field("requester_id").
			Required().
			Unique(),
		edge.To("reviewer", User.Type).
			Field("reviewer_id").
			Required().
			Unique(),
	}
}
