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

// Fields of the Retrospective.
func (Retrospective) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.String("document_name"),
		field.Enum("type").Values("quick", "full"),
		field.Enum("state").Values("draft", "in_review", "meeting", "closed"),
	}
}

// Edges of the Retrospective.
func (Retrospective) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("incident", Incident.Type).
			Unique().Required().Field("incident_id"),
		edge.From("discussions", RetrospectiveDiscussion.Type).
			Ref("retrospective"),
		edge.From("system_analysis", SystemAnalysis.Type).
			Ref("retrospective"),
	}
}
