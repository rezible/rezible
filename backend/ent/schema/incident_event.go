package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// IncidentEvent holds the schema definition for the IncidentEvent entity.
type IncidentEvent struct {
	ent.Schema
}

func (IncidentEvent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

// Fields of the IncidentEvent.
func (IncidentEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.Time("timestamp"),
		field.Enum("kind").
			Values("observation", "action", "decision", "context"),
		field.String("title").
			NotEmpty(),
		field.Text("description").
			Optional(),
		field.Bool("is_key").Default(false),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.UUID("created_by", uuid.UUID{}),
		field.Int("sequence").Default(0),
		field.Bool("is_draft").Default(false),
	}
}

// Indexes of the IncidentEvent.
func (IncidentEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("incident_id", "timestamp"),
		index.Fields("incident_id", "timestamp", "sequence").Unique(),
	}
}

// Edges of the IncidentEvent.
func (IncidentEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incident", Incident.Type).
			Ref("events").
			Field("incident_id").
			Unique().
			Required(),
		edge.To("context", IncidentEventContext.Type).
			Unique(),
		edge.To("factors", IncidentEventContributingFactor.Type),
		edge.To("evidence", IncidentEventEvidence.Type),
		edge.To("system_components", SystemComponent.Type).
			Through("event_components", IncidentEventSystemComponent.Type),
	}
}

type IncidentEventSystemComponent struct {
	ent.Schema
}

func (IncidentEventSystemComponent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (IncidentEventSystemComponent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("incident_event_id", uuid.UUID{}),
		field.UUID("system_component_id", uuid.UUID{}),
		field.Enum("relationship").
			Values("primary", "affected", "contributing"),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (IncidentEventSystemComponent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", IncidentEventSystemComponent.Type).Unique().Required().Field("incident_event_id"),
		edge.To("system_component", SystemComponent.Type).Unique().Required().Field("system_component_id"),
	}
}

type IncidentEventContext struct {
	ent.Schema
}

func (IncidentEventContext) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (IncidentEventContext) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("system_state").Optional(),
		field.JSON("decision_options", []string{}).Optional(),
		field.Text("decision_rationale").Optional(),
		field.JSON("involved_personnel", []string{}).Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (IncidentEventContext) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", IncidentEvent.Type).Ref("context").Unique().Required(),
	}
}

type IncidentEventContributingFactor struct {
	ent.Schema
}

func (IncidentEventContributingFactor) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (IncidentEventContributingFactor) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("factor_type").NotEmpty(),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (IncidentEventContributingFactor) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", IncidentEvent.Type).Ref("factors").Unique().Required(),
	}
}

type IncidentEventEvidence struct {
	ent.Schema
}

func (IncidentEventEvidence) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (IncidentEventEvidence) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Enum("evidence_type").Values("log", "metric", "chat", "ticket", "other"),
		field.String("url").NotEmpty(),
		field.String("title").NotEmpty(),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (IncidentEventEvidence) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", IncidentEvent.Type).Ref("evidence").Unique().Required(),
	}
}
