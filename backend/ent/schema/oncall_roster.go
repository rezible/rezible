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

func (OncallRoster) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		ArchiveMixin{},
	}
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

		edge.From("events", OncallEvent.Type).Ref("roster"),
		edge.From("annotations", OncallAnnotation.Type).Ref("roster"),

		edge.From("teams", Team.Type).Ref("oncall_rosters"),
		edge.From("shifts", OncallShift.Type).Ref("roster"),

		edge.From("user_watchers", User.Type).Ref("watched_oncall_rosters"),

		edge.From("metrics", OncallRosterMetrics.Type).Ref("roster"),
	}
}

type OncallRosterMetrics struct {
	ent.Schema
}

func (OncallRosterMetrics) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (OncallRosterMetrics) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("roster_id", uuid.UUID{}),
	}
}

func (OncallRosterMetrics) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("roster", OncallRoster.Type).
			Unique().Required().Field("roster_id"),
	}
}
