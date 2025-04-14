package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("name"),
		field.String("email"),
		field.String("chat_id").Optional(),
		field.String("timezone").Optional(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("teams", Team.Type).Ref("users"),
		//edge.From("ladders", Ladder.Type).Ref("users"),

		edge.To("watched_oncall_rosters", OncallRoster.Type),

		edge.From("oncall_schedules", OncallScheduleParticipant.Type).Ref("user"),
		edge.From("oncall_shifts", OncallUserShift.Type).Ref("user"),
		edge.From("oncall_shift_covers", OncallUserShiftCover.Type).Ref("user"),
		edge.From("oncall_annotations", OncallAnnotation.Type).Ref("creator"),

		edge.From("incident_role_assignments", IncidentRoleAssignment.Type).Ref("user"),
		edge.To("incident_debriefs", IncidentDebrief.Type),

		edge.To("assigned_tasks", Task.Type),
		edge.To("created_tasks", Task.Type),

		edge.From("retrospective_review_requests", RetrospectiveReview.Type).Ref("requester"),
		edge.From("retrospective_review_responses", RetrospectiveReview.Type).Ref("reviewer"),
	}
}
func (User) Indexes() []ent.Index {
	return []ent.Index{}
}
