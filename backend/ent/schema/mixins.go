package schema

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	gen "github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/hook"
	"github.com/rezible/rezible/ent/intercept"
	"github.com/rezible/rezible/ent/privacy"
	"github.com/rezible/rezible/ent/schema/privacyrules"
)

type BaseMixin struct {
	mixin.Schema
}

func (BaseMixin) Policy() ent.Policy {
	return privacy.Policy{
		Query: privacy.QueryPolicy{
			privacyrules.DenyIfNoAccessContext(),
			privacyrules.AllowIfSystemRole(),
		},
		Mutation: privacy.MutationPolicy{
			privacyrules.DenyIfNoAccessContext(),
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
	return privacyrules.FilterTenantRule()
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

const archiveKey = "include_archived"

func IncludeArchived(parent context.Context) context.Context {
	return context.WithValue(parent, archiveKey, true)
}

func (d ArchiveMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			// Skip archive, means include archived entities.
			if skip, _ := ctx.Value(archiveKey).(bool); skip {
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
					// Skip archival, means delete the entity permanently.
					if skip, _ := ctx.Value(archiveKey).(bool); skip {
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
