package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// OncallRoster holds the schema definition for the OncallRoster entity.
type OncallRoster struct {
	ent.Schema
}

// Fields of the OncallRoster.
func (OncallRoster) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name"),
		field.String("slug").Unique(),
		field.String("provider_id").Unique(),
		field.String("timezone").Optional(),
		field.String("chat_handle").Optional(),
		field.String("chat_channel_id").Optional(),
		field.UUID("handover_template_id", uuid.UUID{}).Optional(),
		//field.UUID("parent_id", uuid.UUID{}).Optional(),
	}
}

// Mixin of the OncallRoster.
func (OncallRoster) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ArchiveMixin{},
	}
}

func (OncallRoster) Indexes() []ent.Index {
	return []ent.Index{
		// index.Fields("id", "parent_id").Unique(),
	}
}

// Edges of the OncallRoster.
func (OncallRoster) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("schedules", OncallSchedule.Type),
		edge.From("handover_template", OncallHandoverTemplate.Type).
			Ref("roster").Unique().Field("handover_template_id"),

		edge.From("teams", Team.Type).Ref("oncall_rosters"),
		edge.From("shifts", OncallUserShift.Type).Ref("roster"),
		edge.From("alerts", OncallAlert.Type).Ref("roster"),
		//edge.To("parent_roster", OncallRoster.Type).Unique().Field("parent_id"),
		//edge.From("children", OncallRoster.Type).Ref("parent_roster"),
	}
}