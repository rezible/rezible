package schema

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type VideoConference struct {
	ent.Schema
}

func (VideoConference) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (VideoConference) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("meeting_session_id", uuid.UUID{}).Optional().Nillable(),
		field.String("provider"),
		field.String("external_id").Optional(),
		field.String("join_url"),
		field.String("host_url").Optional(),
		field.String("dial_in").Optional(),
		field.String("passcode").Optional(),
		field.Enum("status").Values("creating", "active", "ended", "failed").Default("creating"),
		field.JSON("metadata", json.RawMessage{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.String("created_by_integration").Optional(),
	}
}

func (VideoConference) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incident", Incident.Type).
			Ref("video_conferences").
			Unique().
			Field("incident_id"),
		edge.From("meeting_session", MeetingSession.Type).
			Ref("video_conference").
			Unique().
			Field("meeting_session_id"),
	}
}

func (VideoConference) Indexes() []ent.Index {
	return []ent.Index{
		//index.Fields("provider", "external_id").Unique(),
		index.Fields("incident_id", "status"),
		index.Fields("meeting_session_id", "status"),
	}
}
