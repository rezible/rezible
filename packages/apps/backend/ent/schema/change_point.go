package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ChangePoint is a provider-reported change at a point in time, such as a deployment or a merged change. Its
// events share one identity; the row points at the latest of them, whose snapshot is the change point's record.
type ChangePoint struct {
	ent.Schema
}

func (ChangePoint) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
		ProviderResourceReferenceMixin{},
	}
}

func (ChangePoint) Fields() []ent.Field {
	return []ent.Field{
		field.String("kind").
			NotEmpty().
			Immutable().
			Comment("The normalized event kind of its events, for example deployment or code_change."),
		field.Time("occurred_at").
			Comment("The change point time: a deployment's latest event time, or a change's merge time."),
		field.UUID("latest_event_id", uuid.UUID{}),
	}
}

func (ChangePoint) Edges() []ent.Edge {
	return []ent.Edge{
		// A change point records its latest event, so it goes with it.
		edge.To("latest_event", NormalizedEvent.Type).
			Field("latest_event_id").
			Required().
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("involvement", ChangePointInvolvement.Type).
			Ref("change_point"),
	}
}

func (ChangePoint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "kind", "provider", "provider_namespace", "provider_resource_ref").
			Unique(),
		index.Fields("tenant_id", "kind", "occurred_at"),
	}
}

// ChangePointInvolvement is one entity a change point's latest event involves: a copy of that event's involvement
// evidence on the entity's alias. A change point's rows are replaced when its latest event changes.
type ChangePointInvolvement struct {
	ent.Schema
}

func (ChangePointInvolvement) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
	}
}

func (ChangePointInvolvement) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("change_point_id", uuid.UUID{}),
		field.String("role").
			NotEmpty().
			Comment("The involvement evidence's assertion, for example deployment_service."),
		field.UUID("subject_alias_id", uuid.UUID{}),
		field.UUID("evidence_id", uuid.UUID{}).
			Comment("The involvement evidence the row copies: what to cite."),
	}
}

func (ChangePointInvolvement) Edges() []ent.Edge {
	return []ent.Edge{
		// An involvement row goes with the change point it belongs to and the evidence it copies.
		edge.To("change_point", ChangePoint.Type).
			Field("change_point_id").
			Required().
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("subject_alias", KnowledgeSubjectAlias.Type).
			Field("subject_alias_id").
			Required().
			Unique(),
		edge.To("evidence", KnowledgeEvidence.Type).
			Field("evidence_id").
			Required().
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (ChangePointInvolvement) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("change_point_id", "subject_alias_id").
			Unique(),
		index.Fields("tenant_id", "subject_alias_id", "role"),
	}
}
