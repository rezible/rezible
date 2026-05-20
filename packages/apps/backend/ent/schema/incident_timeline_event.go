package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// IncidentTimelineEvent holds the schema definition for the IncidentTimelineEvent entity.
type IncidentTimelineEvent struct {
	ent.Schema
}

func (IncidentTimelineEvent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

var incidentEventKinds = []string{"observation", "context", "decision", "action"}

// Fields of the IncidentTimelineEvent.
func (IncidentTimelineEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.UUID("event_id", uuid.UUID{}).Optional(),
		field.Time("timestamp"),
		field.Enum("kind").Values(incidentEventKinds...),
		field.String("title").NotEmpty(),
		field.Text("description").Optional(),
		field.Bool("is_key").Default(false),
		field.Int("sequence").Default(0),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Indexes of the IncidentTimelineEvent.
func (IncidentTimelineEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("kind"),
	}
}

// Edges of the IncidentTimelineEvent.
func (IncidentTimelineEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incident", Incident.Type).
			Ref("timeline_events").
			Field("incident_id").
			Unique().
			Required(),
		edge.To("event", NormalizedEvent.Type).
			Unique().
			Field("event_id"),
		edge.To("context", IncidentTimelineEventContext.Type).
			Unique(),
		edge.To("factors", IncidentTimelineEventContributingFactor.Type),
		edge.To("evidence", IncidentTimelineEventEvidence.Type),
		edge.From("topology_context", IncidentTimelineEventTopologyContext.Type).
			Ref("event"),
	}
}

type IncidentTimelineEventTopologyContext struct {
	ent.Schema
}

func (IncidentTimelineEventTopologyContext) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (IncidentTimelineEventTopologyContext) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("incident_event_id", uuid.UUID{}),
		field.UUID("knowledge_entity_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("snapshot_entity_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("relationship").
			Values("primary", "affected", "contributing"),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (IncidentTimelineEventTopologyContext) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", IncidentTimelineEvent.Type).
			Unique().
			Required().
			Field("incident_event_id"),
		edge.To("knowledge_entity", KnowledgeEntity.Type).
			Unique().
			Field("knowledge_entity_id"),
		edge.To("snapshot_entity", SystemTopologySnapshotEntity.Type).
			Unique().
			Field("snapshot_entity_id"),
	}
}

type IncidentTimelineEventContext struct {
	ent.Schema
}

func (IncidentTimelineEventContext) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (IncidentTimelineEventContext) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("system_state").Optional(),
		field.JSON("decision_options", []string{}).Optional(),
		field.Text("decision_rationale").Optional(),
		field.JSON("involved_personnel", []string{}).Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (IncidentTimelineEventContext) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", IncidentTimelineEvent.Type).
			Ref("context").
			Unique().
			Required(),
	}
}

type IncidentTimelineEventContributingFactor struct {
	ent.Schema
}

func (IncidentTimelineEventContributingFactor) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (IncidentTimelineEventContributingFactor) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("factor_type").NotEmpty(),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (IncidentTimelineEventContributingFactor) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", IncidentTimelineEvent.Type).
			Ref("factors").
			Unique().
			Required(),
	}
}

type IncidentTimelineEventEvidence struct {
	ent.Schema
}

func (IncidentTimelineEventEvidence) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (IncidentTimelineEventEvidence) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Enum("evidence_type").Values("log", "metric", "chat", "ticket", "other"),
		field.String("url").NotEmpty(),
		field.String("title").NotEmpty(),
		field.Text("description").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (IncidentTimelineEventEvidence) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", IncidentTimelineEvent.Type).
			Ref("evidence").
			Unique().
			Required(),
	}
}
