package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

type IncidentDebrief struct {
	ent.Schema
}

func (IncidentDebrief) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (IncidentDebrief) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Bool("required"),
		field.Bool("started"),
	}
}

func (IncidentDebrief) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incident", Incident.Type).
			Ref("debriefs").Unique().Required().Field("incident_id"),
		edge.From("user", User.Type).
			Ref("incident_debriefs").Unique().Required().Field("user_id"),
		edge.To("messages", IncidentDebriefMessage.Type),
		edge.To("suggestions", IncidentDebriefSuggestion.Type),
	}
}

type IncidentDebriefMessage struct {
	ent.Schema
}

func (IncidentDebriefMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (IncidentDebriefMessage) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("debrief_id", uuid.UUID{}),
		field.UUID("question_id", uuid.UUID{}).Optional(),
		field.Time("created_at").Default(time.Now),
		field.Enum("type").Values("user", "assistant", "question"),
		field.Enum("requested_tool").Values("rating").Optional(),
		field.Text("body"),
	}
}

func (IncidentDebriefMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("debrief", IncidentDebrief.Type).
			Ref("messages").Unique().Required().Field("debrief_id"),
		edge.To("from_question", IncidentDebriefQuestion.Type).Unique().
			Field("question_id"),
	}
}

type IncidentDebriefQuestion struct {
	ent.Schema
}

func (IncidentDebriefQuestion) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields of the IncidentDebriefQuestion.
func (IncidentDebriefQuestion) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.Text("content"),
	}
}

// Edges of the IncidentDebriefQuestion.
func (IncidentDebriefQuestion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("messages", IncidentDebriefMessage.Type).Ref("from_question"),

		edge.To("incident_fields", IncidentField.Type),
		edge.To("incident_roles", IncidentRole.Type),
		edge.To("incident_severities", IncidentSeverity.Type),
		edge.To("incident_tags", IncidentTag.Type),
		edge.To("incident_types", IncidentType.Type),
	}
}

type IncidentDebriefSuggestion struct {
	ent.Schema
}

func (IncidentDebriefSuggestion) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields of the IncidentDebriefSuggestion.
func (IncidentDebriefSuggestion) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.Text("content"),
	}
}

// Edges of the IncidentDebriefSuggestion.
func (IncidentDebriefSuggestion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("debrief", IncidentDebrief.Type).Ref("suggestions").Required().Unique(),
	}
}
