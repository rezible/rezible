package schema

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"

	gen "github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/hook"
	"github.com/rezible/rezible/ent/intercept"
	"github.com/rezible/rezible/ent/privacy"
	"github.com/rezible/rezible/ent/schema/rules"
)

type BaseMixin struct {
	mixin.Schema
}

func (BaseMixin) Policy() ent.Policy {
	return privacy.Policy{
		Query: privacy.QueryPolicy{
			rules.DenyIfNoAccessContext(),
			rules.AllowIfSystemRole(),
		},
		Mutation: privacy.MutationPolicy{
			rules.DenyIfNoAccessContext(),
			rules.DenyIfAnonymous(),
		},
	}
}

type TenantMixin struct {
	mixin.Schema
}

func (TenantMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Int("tenant_id").Immutable(),
	}
}

func (TenantMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id"),
	}
}

func (TenantMixin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("tenant", Tenant.Type).
			Field("tenant_id").
			Unique().
			Required().
			Immutable(),
	}
}

func (TenantMixin) Policy() ent.Policy {
	return rules.FilterTenantRule()
}

// ArchiveMixin implements the soft delete pattern for schemas.
type ArchiveMixin struct {
	mixin.Schema
}

func (ArchiveMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("archive_time").
			Optional(),
	}
}

const includeArchivedContextKey = "include_archived"

func IncludeArchived(parent context.Context) context.Context {
	return context.WithValue(parent, includeArchivedContextKey, true)
}

func (d ArchiveMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			// Skip archive, means include archived entities.
			if skip, _ := ctx.Value(includeArchivedContextKey).(bool); skip {
				return nil
			}
			d.P(q)
			return nil
		}),
	}
}

func (d ArchiveMixin) Hooks() []ent.Hook {
	type ArchiveableMutation interface {
		SetOp(ent.Op)
		Client() *gen.Client
		SetArchiveTime(time.Time)
		WhereP(...func(*sql.Selector))
	}
	return []ent.Hook{
		hook.On(
			func(next ent.Mutator) ent.Mutator {
				return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
					if skip, _ := ctx.Value(includeArchivedContextKey).(bool); skip {
						return next.Mutate(ctx, m)
					}
					mx, ok := m.(ArchiveableMutation)
					if !ok {
						return nil, fmt.Errorf("unexpected mutation type %T (not archivable)", m)
					}
					d.P(mx)
					mx.SetOp(ent.OpUpdate)
					mx.SetArchiveTime(time.Now())
					return mx.Client().Mutate(ctx, m)
				})
			},
			ent.OpDeleteOne|ent.OpDelete,
		),
	}
}

func (d ArchiveMixin) P(w interface{ WhereP(...func(*sql.Selector)) }) {
	w.WhereP(
		sql.FieldIsNull(d.Fields()[0].Descriptor().Name),
	)
}

type IntegrationMixin struct {
	mixin.Schema
	Required bool
}

func (i IntegrationMixin) Fields() []ent.Field {
	idField := field.String("external_id")
	if !i.Required {
		idField.Optional()
	} else {
		idField.NotEmpty()
	}
	return []ent.Field{idField}
}
