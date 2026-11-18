package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent/schema/annotations"
)

// ObservationScope is what an observation is about: the runtime knowledge entities, such as services, that its
// evidence points to. An observation is evidence that something may be wrong; today it is an alert episode.
//
// Situations group observations by their scope. The scope is computed once per observation and recomputed when the
// observation changes. An empty scope means the observation points to no known entity.
type ObservationScope struct {
	ent.Schema
}

func (ObservationScope) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
	}
}

func (ObservationScope) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Checks: map[string]string{
			"observation_scope_revision_nonnegative": "revision >= 0",
			// Each observation kind has its own reference column, and exactly one is set.
			"observation_scope_exactly_one_reference": "num_nonnulls(alert_episode_id) = 1",
		}},
	}
}

func (ObservationScope) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("alert_episode_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Int("revision").
			NonNegative().
			Comment("The observation revision the scope was computed from; never lowered"),
		field.Bool("broad").
			Comment("Whether the observation points to too many entities to group other observations with it"),
	}
}

func (ObservationScope) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("alert_episode", AlertEpisode.Type).
			Ref("observation_scope").
			Field("alert_episode_id").
			Unique(),
		edge.To("entities", ObservationScopeEntity.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// ObservationScopeEntity is one entity in an observation's scope.
type ObservationScopeEntity struct {
	ent.Schema
}

func (ObservationScopeEntity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
		TenantMixin{},
	}
}

func (ObservationScopeEntity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("scope_id", uuid.UUID{}).Immutable(),
		field.UUID("knowledge_entity_id", uuid.UUID{}),
	}
}

func (ObservationScopeEntity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("scope", ObservationScope.Type).
			Ref("entities").
			Field("scope_id").
			Required().
			Unique().
			Immutable(),
		edge.To("knowledge_entity", KnowledgeEntity.Type).
			Field("knowledge_entity_id").
			Required().
			Unique().
			Annotations(annotations.EntityMergeDeletesDuplicateRows()),
	}
}

func (ObservationScopeEntity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "scope_id", "knowledge_entity_id").Unique(),
		index.Fields("tenant_id", "knowledge_entity_id"),
	}
}
