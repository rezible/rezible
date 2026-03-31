package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// IncidentMilestone holds the schema definition for the IncidentMilestone entity.
type IncidentMilestone struct {
	ent.Schema
}

func (IncidentMilestone) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

var incidentMilestoneKinds = []string{"impact", "detected", "opened", "mitigation", "resolution"}

// Fields of the IncidentMilestone.
func (IncidentMilestone) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("kind").Values(incidentMilestoneKinds...),
		field.Time("timestamp"),
		field.String("description").Optional(),
		field.String("source").Optional(),
		field.JSON("metadata", map[string]string{}).Optional(),
	}
}

// Edges of the IncidentEvent.
func (IncidentMilestone) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incident", Incident.Type).
			Ref("milestones").Unique().Required().Field("incident_id"),
		edge.From("user", User.Type).
			Ref("incident_milestones").Unique().Required().Field("user_id"),
	}
}
