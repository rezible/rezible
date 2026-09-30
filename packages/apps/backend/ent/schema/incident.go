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
		TimestampsMixin{},
		KnowledgeEntityLinkMixin{},
	}
}

var incidentResponseStates = []string{"unknown", "started", "mitigated", "resolved"}

// Fields of the Incident.
func (Incident) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("slug").Unique(),
		field.String("title"),
		field.UUID("severity_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("type_id", uuid.UUID{}).Optional().Nillable(),
		field.String("summary").Optional(),
		field.String("chat_channel_id").Optional(),
		field.Enum("response_state").Values(incidentResponseStates...).Default("unknown"),
		field.Time("opened_at").Default(time.Now),
		field.Time("resolved_at").Optional().Nillable(),
	}
}

// Edges of the Incident.
func (Incident) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("severity", IncidentSeverity.Type).
			Unique().
			Field("severity_id"),
		edge.To("type", IncidentType.Type).
			Unique().
			Field("type_id"),

		edge.To("milestones", IncidentMilestone.Type),

		edge.To("retrospective", Retrospective.Type).
			Unique(),

		edge.From("role_assignments", IncidentRoleAssignment.Type).
			Ref("incident"),

		edge.To("linked_incidents", Incident.Type).
			Through("incident_links", IncidentLink.Type),

		edge.To("situations", Situation.Type),

		edge.To("field_selections", IncidentFieldOption.Type),
		edge.To("tag_assignments", IncidentTag.Type),
		edge.From("impacts", IncidentImpact.Type).
			Ref("incident"),
		edge.To("debriefs", IncidentDebrief.Type),
		edge.From("video_conferences", VideoConference.Type).
			Ref("incident"),
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
		field.Enum("link_type").Values("parent", "child", "similar"),
		field.String("description").Optional(),
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
