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

func (Team) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		IntegrationMixin{},
	}
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
		edge.From("users", User.Type).Ref("teams").
			Through("team_memberships", TeamMembership.Type),

		edge.To("oncall_rosters", OncallRoster.Type),

		edge.From("scheduled_meetings", MeetingSchedule.Type).Ref("owning_team"),
	}
}
