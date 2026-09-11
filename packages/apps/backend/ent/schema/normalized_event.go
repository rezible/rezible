package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type NormalizedEvent struct {
	ent.Schema
}

func (NormalizedEvent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		ProviderResourceReferenceMixin{},
	}
}

func (NormalizedEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("integration_id", uuid.UUID{}).Optional().Nillable(),
		field.String("kind").Immutable().NotEmpty().
			Comment("Normalized kind of the primary subject this event is about."),
		field.String("provider_event_source").Immutable().NotEmpty().
			Comment("Provider-specific event stream or webhook source the event came from."),
		field.String("provider_event_ref").Immutable().NotEmpty().
			Comment("Stable provider reference for the source event, used with the provider fields for idempotency."),
		field.Bytes("attributes").Immutable().
			Comment("Normalized JSON attributes for this event kind."),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("occurred_at").Immutable(),
		field.Time("received_at").Immutable(),
	}
}

func (NormalizedEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("integration", Integration.Type).
			Unique().
			Field("integration_id").
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("projection", NormalizedEventProjection.Type).Unique(),

		edge.From("situation_observation_groups", SituationObservationGroup.Type).Ref("events"),
		edge.From("analysis_entry_subjects", SystemAnalysisEntrySubject.Type).Ref("normalized_event"),
	}
}

func (NormalizedEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "provider", "provider_namespace", "provider_event_source", "provider_event_ref", "provider_resource_ref").
			Unique(),
		index.Fields("tenant_id", "provider", "provider_namespace", "provider_event_source", "occurred_at"),
	}
}

type NormalizedEventProjection struct {
	ent.Schema
}

func (NormalizedEventProjection) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (NormalizedEventProjection) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("event_id", uuid.UUID{}).Immutable(),
		field.Time("completed_at").Default(time.Now).Immutable(),
	}
}

func (NormalizedEventProjection) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", NormalizedEvent.Type).
			Immutable().
			Required().
			Unique().
			Field("event_id"),
		edge.From("projection_entities", NormalizedEventProjectionEntity.Type).
			Ref("projection"),
	}
}

type NormalizedEventProjectionEntity struct {
	ent.Schema
}

func (NormalizedEventProjectionEntity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (NormalizedEventProjectionEntity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("projection_id", uuid.UUID{}),
		field.String("domain_entity_kind"),
		field.UUID("domain_entity_id", uuid.UUID{}),
	}
}

func (NormalizedEventProjectionEntity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("projection", NormalizedEventProjection.Type).
			Unique().
			Required().
			Field("projection_id"),
	}
}

func (NormalizedEventProjectionEntity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "projection_id", "domain_entity_id").Unique(),
		index.Fields("tenant_id", "domain_entity_kind"),
	}
}

type EventAnnotation struct {
	ent.Schema
}

func (EventAnnotation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
	}
}

func (EventAnnotation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("event_id", uuid.UUID{}),
		field.UUID("creator_id", uuid.UUID{}),
		field.Time("created_at").Default(time.Now),
		field.Int("minutes_occupied"),
		field.Text("notes"),
		field.JSON("tags", []string{}),
	}
}

func (EventAnnotation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", NormalizedEvent.Type).
			Unique().
			Required().
			Field("event_id"),
		edge.To("creator", User.Type).
			Unique().
			Required().
			Field("creator_id"),

		edge.From("handovers", OncallShiftHandover.Type).
			Ref("pinned_annotations"),
	}
}
