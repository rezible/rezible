package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("email"),
		field.String("name").Default(""),
		field.String("chat_id").Optional(),
		field.String("timezone").Optional(),
		field.String("auth_provider_id").Optional(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("organization_role", OrganizationRole.Type).Unique(),

		edge.To("teams", Team.Type).
			Through("team_memberships", TeamMembership.Type),

		edge.To("watched_oncall_rosters", OncallRoster.Type),

		edge.From("oncall_schedules", OncallScheduleParticipant.Type).Ref("user"),
		edge.From("oncall_shifts", OncallShift.Type).Ref("user"),

		edge.From("event_annotations", EventAnnotation.Type).Ref("creator"),

		edge.From("integration_oauth_states", IntegrationOAuthState.Type).Ref("user"),

		edge.To("incidents", Incident.Type).
			Through("role_assignments", IncidentRoleAssignment.Type),
		edge.To("incident_milestones", IncidentMilestone.Type),

		edge.To("incident_debriefs", IncidentDebrief.Type),

		edge.To("assigned_tasks", Task.Type),
		edge.To("created_tasks", Task.Type),

		edge.From("retrospective_review_requests", RetrospectiveReview.Type).Ref("requester"),
		edge.From("retrospective_review_responses", RetrospectiveReview.Type).Ref("reviewer"),
		edge.From("retrospective_comments", RetrospectiveComment.Type).Ref("user"),

		edge.From("document_accesses", DocumentAccess.Type).Ref("user"),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("auth_provider_id").Unique(),
	}
}
