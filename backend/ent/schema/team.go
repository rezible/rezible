package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Team holds the schema definition for the Team entity.
type Team struct {
	ent.Schema
}

// Fields of the Team.
func (Team) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("slug").Unique(),
		field.String("name"),
		field.String("chat_channel_id").Optional(),
		field.String("timezone").Optional(),
	}
}

// Edges of the Team.
func (Team) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("users", User.Type),

		edge.From("services", Service.Type).Ref("owner_team"),

		//edge.From("ladders", Ladder.Type).Ref("teams"),
		edge.To("oncall_rosters", OncallRoster.Type),

		edge.From("subscriptions", Subscription.Type).
			Ref("team"),
		edge.From("incident_assignments", IncidentTeamAssignment.Type).
			Ref("team"),
		edge.From("scheduled_meetings", MeetingSchedule.Type).Ref("owning_team"),
	}
}
