package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// RetrospectiveDiscussion holds the schema definition for the RetrospectiveDiscussion entity.
type RetrospectiveDiscussion struct {
	ent.Schema
}

// Fields of the RetrospectiveDiscussion.
func (RetrospectiveDiscussion) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("retrospective_id", uuid.UUID{}),
		field.Bytes("content"),
	}
}

// Edges of the RetrospectiveDiscussion.
func (RetrospectiveDiscussion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("retrospective", Retrospective.Type).
			Field("retrospective_id").
			Required().
			Unique(),
		edge.From("replies", RetrospectiveDiscussionReply.Type).
			Ref("discussion"),
		edge.From("review", RetrospectiveReview.Type).
			Ref("discussion"),
	}
}

// RetrospectiveDiscussionReply holds the schema definition for the RetrospectiveDiscussionReply entity.
type RetrospectiveDiscussionReply struct {
	ent.Schema
}

// Fields of the RetrospectiveDiscussionReply.
func (RetrospectiveDiscussionReply) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.Bytes("content"),
	}
}

// Edges of the RetrospectiveDiscussionReply.
func (RetrospectiveDiscussionReply) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("discussion", RetrospectiveDiscussion.Type).
			Required().
			Unique(),
		edge.To("replies", RetrospectiveDiscussionReply.Type).
			From("parent_reply").
			Unique(),
	}
}
