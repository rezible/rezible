package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Document struct {
	ent.Schema
}

func (Document) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AccessScopeMixin{},
		TenantMixin{},
	}
}

func (Document) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.Bytes("content"),
		field.Bool("access_restricted").Default(false),
	}
}

func (Document) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("accesses", DocumentAccess.Type).Ref("document"),
		edge.To("retrospective", Retrospective.Type).Unique(),
	}
}

type DocumentAccess struct {
	ent.Schema
}

func (DocumentAccess) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AccessScopeMixin{},
		TenantMixin{},
		TimestampsMixin{},
	}
}

func (DocumentAccess) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("document_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}).Optional(),
		field.UUID("team_id", uuid.UUID{}).Optional(),
		field.Bool("can_view").Default(false),
		field.Bool("can_edit").Default(false),
		field.Bool("can_manage").Default(false),
	}
}

func (DocumentAccess) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("document", Document.Type).Unique().Required().Field("document_id"),
		edge.To("user", User.Type).Unique().Field("user_id"),
		edge.To("team", Team.Type).Unique().Field("team_id"),
	}
}
