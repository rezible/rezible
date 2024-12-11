package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// IncidentDebriefQuestion holds the schema definition for the IncidentDebriefQuestion entity.
type IncidentDebriefQuestion struct {
	ent.Schema
}

// Fields of the IncidentDebriefQuestion.
func (IncidentDebriefQuestion) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.Text("content"),
	}
}

// Edges of the IncidentDebriefQuestion.
func (IncidentDebriefQuestion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("messages", IncidentDebriefMessage.Type).Ref("from_question"),

		edge.To("incident_fields", IncidentField.Type),
		edge.To("incident_roles", IncidentRole.Type),
		edge.To("incident_severities", IncidentSeverity.Type),
		edge.To("incident_tags", IncidentTag.Type),
		edge.To("incident_types", IncidentType.Type),
	}
}
