package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Incident holds the schema definition for the Incident entity.
type Incident struct {
	ent.Schema
}

// Fields of the Incident.
func (Incident) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("slug").Unique(),
		field.String("title"),
		field.Bool("private").Default(false),
		field.String("summary"),
		field.Time("opened_at"),
		field.Time("modified_at"),
		field.Time("closed_at"),
		field.String("provider_id"),
		field.String("chat_channel_id").Optional(),
		field.UUID("severity_id", uuid.UUID{}).Optional(),
		field.UUID("type_id", uuid.UUID{}).Optional(),
		// field.UUID("retrospective_id", uuid.UUID{}),
		field.UUID("analysis_id", uuid.UUID{}),
	}
}

// Edges of the Incident.
func (Incident) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("environments", Environment.Type),
		edge.To("severity", IncidentSeverity.Type).
			Unique().Field("severity_id"),
		edge.To("type", IncidentType.Type).
			Unique().Field("type_id"),

		edge.From("team_assignments", IncidentTeamAssignment.Type).
			Ref("incident"),
		edge.From("role_assignments", IncidentRoleAssignment.Type).
			Ref("incident"),

		edge.To("retrospective", Retrospective.Type),
		edge.To("milestones", IncidentMilestone.Type),
		edge.To("events", IncidentEvent.Type),
		edge.From("system_analysis", SystemAnalysis.Type).
			Ref("incident"),

		edge.To("linked_incidents", Incident.Type).
			Through("incident_links", IncidentLink.Type),

		edge.To("field_selections", IncidentFieldOption.Type),
		edge.To("tasks", Task.Type),
		edge.To("tag_assignments", IncidentTag.Type),
		edge.To("debriefs", IncidentDebrief.Type),
		edge.To("review_sessions", MeetingSession.Type),
	}
}
