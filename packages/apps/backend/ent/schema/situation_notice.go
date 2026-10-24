package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// SituationNotice is the chat message announcing a situation in one installation's channel. It is posted
// once, then edited as the situation changes.
type SituationNotice struct {
	ent.Schema
}

func (SituationNotice) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SituationNotice) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("situation_id", uuid.UUID{}).Immutable(),
		field.UUID("integration_id", uuid.UUID{}).
			Immutable().
			Comment("The installation that posts it"),
		field.String("channel_id").
			NotEmpty().
			Immutable().
			Comment("The installation's notice channel when the notice was created"),
		field.String("channel_name").
			Optional().
			Comment("The channel's name, as the provider last reported it"),
		field.String("message_ref").
			Optional().
			Comment("The provider's identifier of the posted message; empty until posted"),
		field.String("permalink").Optional(),
		field.Enum("status").
			Values("pending", "posted", "failed", "skipped").
			Default("pending"),
		field.String("error").
			Optional().
			Comment("Why the last delivery failed, or why the notice was skipped"),
		field.Time("last_attempt_at").Optional().Nillable(),
		field.Time("first_posted_at").
			Optional().
			Nillable().
			Comment("When the message was first posted; the hourly notice limit counts these"),
	}
}

func (SituationNotice) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("situation", Situation.Type).
			Ref("notices").
			Field("situation_id").
			Required().
			Unique().
			Immutable(),
		edge.To("integration", Integration.Type).
			Field("integration_id").
			Required().
			Unique().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (SituationNotice) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "situation_id", "integration_id").Unique(),
		index.Fields("integration_id", "channel_id", "first_posted_at"),
	}
}

// SituationNoticeRating is one chat user's rating of a situation's notice. A second rating by the same
// user replaces the first.
type SituationNoticeRating struct {
	ent.Schema
}

func (SituationNoticeRating) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (SituationNoticeRating) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("situation_id", uuid.UUID{}).Immutable(),
		field.String("chat_user_id").
			NotEmpty().
			Immutable().
			Comment("The rating person's chat user ID"),
		field.UUID("user_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("The rating person's Rezible user, when the chat user maps to one"),
		field.Bool("useful"),
	}
}

func (SituationNoticeRating) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("situation", Situation.Type).
			Ref("notice_ratings").
			Field("situation_id").
			Required().
			Unique().
			Immutable(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique(),
	}
}

func (SituationNoticeRating) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "situation_id", "chat_user_id").Unique(),
	}
}
