package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
		IntegrationDataMixin{},
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

		edge.From("document_accesses", DocumentAccess.Type).Ref("team"),
	}
}

type TeamMembership struct {
	ent.Schema
}

func (TeamMembership) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (TeamMembership) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("team_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("role").Values("admin", "member").Default("member"),
	}
}

func (TeamMembership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("team", Team.Type).
			Required().
			Unique().
			Field("team_id"),
		edge.To("user", User.Type).
			Required().
			Unique().
			Field("user_id"),
	}
}

func (TeamMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("team_id", "user_id").Unique(),
	}
}
