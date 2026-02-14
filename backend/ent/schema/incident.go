package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Incident holds the schema definition for the Incident entity.
type Incident struct {
	ent.Schema
}

func (Incident) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		IntegrationMixin{},
		TimestampsMixin{},
	}
}

// Fields of the Incident.
func (Incident) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("slug").Unique(),
		field.String("title"),
		field.UUID("severity_id", uuid.UUID{}),
		field.UUID("type_id", uuid.UUID{}),
		field.String("summary").Optional(),
		field.String("chat_channel_id").Optional(),
		field.Time("opened_at").Default(time.Now),
	}
}

// Edges of the Incident.
func (Incident) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("severity", IncidentSeverity.Type).
			Unique().
			Required().
			Field("severity_id"),
		edge.To("type", IncidentType.Type).
			Unique().
			Required().
			Field("type_id"),

		edge.To("milestones", IncidentMilestone.Type),
		edge.To("events", IncidentEvent.Type),

		edge.To("retrospective", Retrospective.Type).
			Unique(),

		edge.From("users", User.Type).
			Ref("incidents").
			Through("user_roles", IncidentRoleAssignment.Type),

		edge.From("role_assignments", IncidentRoleAssignment.Type).
			Ref("incident"),

		edge.To("linked_incidents", Incident.Type).
			Through("incident_links", IncidentLink.Type),

		edge.To("field_selections", IncidentFieldOption.Type),
		edge.To("tasks", Task.Type),
		edge.To("tag_assignments", IncidentTag.Type),
		edge.To("debriefs", IncidentDebrief.Type),
		edge.To("review_sessions", MeetingSession.Type),
		edge.To("video_conferences", VideoConference.Type),
	}
}

type IncidentLink struct {
	ent.Schema
}

func (IncidentLink) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (IncidentLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("incident_id", uuid.UUID{}),
		field.UUID("linked_incident_id", uuid.UUID{}),
		field.String("description").Optional(),
		field.Enum("link_type").Values("parent", "child", "similar"),
	}
}

func (IncidentLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("incident", Incident.Type).
			Required().
			Unique().
			Field("incident_id"),
		edge.To("linked_incident", Incident.Type).
			Required().
			Unique().
			Field("linked_incident_id"),
	}
}
