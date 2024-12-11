package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

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
