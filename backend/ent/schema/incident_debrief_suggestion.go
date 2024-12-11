package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// IncidentDebriefSuggestion holds the schema definition for the IncidentDebriefSuggestion entity.
type IncidentDebriefSuggestion struct {
	ent.Schema
}

// Fields of the IncidentDebriefSuggestion.
func (IncidentDebriefSuggestion) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.Text("content"),
	}
}

// Edges of the IncidentDebriefSuggestion.
func (IncidentDebriefSuggestion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("debrief", IncidentDebrief.Type).Ref("suggestions").Required().Unique(),
	}
}
