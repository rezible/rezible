package schema

import (
	"entgo.io/ent/schema/edge"
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

var (
	taskKinds  = []string{"cleanup", "detect", "mitigate", "prevent"}
	taskStates = []string{"open", "completed", "cancelled"}
)

// Task holds the schema definition for the Task entity.
type Task struct {
	ent.Schema
}

func (Task) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
		ArchiveMixin{},
	}
}

// Fields of the Task.
func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.Int("version").Default(1),
		field.String("title"),
		field.Text("description").Default(""),
		field.Enum("kind").Values(taskKinds...),
		field.Enum("state").Values(taskStates...).Default("open"),
		field.Time("due_at").Optional().Nillable(),
		field.UUID("incident_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("origin_entry_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("assignee_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("creator_id", uuid.UUID{}).Optional().Nillable(),
	}
}

// Edges of the Task.
func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("tickets", Ticket.Type),
		edge.To("origin_entry", SystemAnalysisEntry.Type).
			Unique().
			Field("origin_entry_id"),

		edge.To("incident", Incident.Type).
			Unique().
			Field("incident_id"),

		edge.To("assignee", User.Type).
			Unique().
			Field("assignee_id"),
		edge.To("creator", User.Type).
			Unique().
			Field("creator_id"),
	}
}

// Ticket holds the schema definition for the Ticket entity.
type Ticket struct {
	ent.Schema
}

func (Ticket) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
		//ProviderResourceReferenceMixin{},
	}
}

// Fields of the Ticket.
func (Ticket) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("title"),
		field.String("reference").Optional().Nillable(),
		field.String("url").Optional().Nillable(),
		field.String("provider").Optional().Nillable(),
		field.String("provider_namespace").Optional().Nillable(),
		field.String("provider_resource_ref").Optional().Nillable(),
	}
}

// Edges of the Ticket.
func (Ticket) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tasks", Task.Type).Ref("tickets"),
	}
}
