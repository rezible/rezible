package schema

import (
	"entgo.io/ent/schema/edge"
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Task holds the schema definition for the Task entity.
type Task struct {
	ent.Schema
}

var (
	taskTypes = []string{
		"cleanup",
		"detect",
		"mitigate",
		"prevent",
	}
)

// Fields of the Task.
func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.Enum("type").Values(taskTypes...),
		field.String("title"),
		field.UUID("incident_id", uuid.UUID{}).Optional(),
		field.UUID("assignee_id", uuid.UUID{}).Optional(),
		field.UUID("creator_id", uuid.UUID{}).Optional(),
		field.String("issue_tracker_id").Optional(),
	}
}

// Edges of the Task.
func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incident", Incident.Type).
			Ref("tasks").Unique().Field("incident_id"),
		edge.From("assignee", User.Type).
			Ref("assigned_tasks").Unique().Field("assignee_id"),
		edge.From("creator", User.Type).
			Ref("created_tasks").Unique().Field("creator_id"),
	}
}
