package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/schema/schematypes"
)

type NormalizedEvent struct {
	ent.Schema
}

func (NormalizedEvent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
		ProviderResourceReferenceMixin{},
	}
}

func (NormalizedEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("integration_id", uuid.UUID{}).Optional().Nillable(),
		field.String("kind").Immutable().NotEmpty().
			Comment("Normalized kind of the primary subject this event is about."),
		field.String("provider_event_source").Immutable().NotEmpty().
			Comment("Provider-specific event stream or webhook source the event came from."),
		field.String("provider_event_ref").Immutable().NotEmpty().
			Comment("Stable provider reference for the source event, used with the provider fields for idempotency."),
		field.Bytes("attributes").
			SchemaType(schemaTypeJsonB).
			Immutable().
			Comment("Normalized JSON attributes for this event kind. Stored as jsonb, so queries filter on them; jsonb keeps neither key order nor whitespace."),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("occurred_at").Immutable(),
		field.Time("last_occurred_at"),
		field.Time("received_at").Immutable(),
		field.Time("failed_at").Optional().Nillable().
			Comment("When work on the event last failed in a way a retry would not fix."),
		field.String("failure_reason").Optional().Nillable().
			Comment("Why work on the event failed, in words safe to show to members."),
	}
}

func (NormalizedEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("integration", Integration.Type).
			Unique().
			Field("integration_id").
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("projection", NormalizedEventProjection.Type).
			Unique(),
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
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
	}
}

func (NormalizedEventProjection) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("event_id", uuid.UUID{}).Immutable(),
		field.Time("completed_at").Default(time.Now).Immutable(),
		field.JSON("skipped_refs", []schematypes.SkippedEvidenceRef{}).Default([]schematypes.SkippedEvidenceRef{}),
	}
}

func (NormalizedEventProjection) Edges() []ent.Edge {
	return []ent.Edge{
		// The inverse of NormalizedEvent.projection: an event's receipt, held by event_id.
		edge.From("event", NormalizedEvent.Type).
			Ref("projection").
			Immutable().
			Required().
			Unique().
			Field("event_id"),
		edge.From("projection_entities", NormalizedEventProjectionEntity.Type).
			Ref("projection"),
	}
}

func (NormalizedEventProjection) Indexes() []ent.Index {
	return []ent.Index{
		// One receipt per event. The one-to-one edge makes event_id unique; under the name of that constraint,
		// this index replaces it with one scoped to the tenant, as receipt lookups are.
		index.Fields("tenant_id", "event_id").
			Unique().
			StorageKey("normalized_event_projections_event_id_key"),
	}
}

type NormalizedEventProjectionEntity struct {
	ent.Schema
}

func (NormalizedEventProjectionEntity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
	}
}

func (NormalizedEventProjectionEntity) Fields() []ent.Field {
	return []ent.Field{
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
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
	}
}

func (EventAnnotation) Fields() []ent.Field {
	return []ent.Field{
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
