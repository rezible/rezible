package schema

import (
	"context"
	"fmt"
	"time"

	gen "github.com/twohundreds/rezible/ent"
	"github.com/twohundreds/rezible/ent/hook"
	"github.com/twohundreds/rezible/ent/intercept"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// ArchiveMixin implements the soft delete pattern for schemas.
type ArchiveMixin struct {
	mixin.Schema
}

// Fields of the ArchiveMixin.
func (ArchiveMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("archive_time").
			Optional(),
	}
}

const archiveKey = "include_archived"

// IncludeArchived returns a new context that skips the archive interceptor/mutators.
func IncludeArchived(parent context.Context) context.Context {
	return context.WithValue(parent, archiveKey, true)
}

// Interceptors of the ArchiveMixin.
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

type ArchiveableMutation interface {
	SetOp(ent.Op)
	Client() *gen.Client
	SetArchiveTime(time.Time)
	WhereP(...func(*sql.Selector))
}

// Hooks of the ArchiveMixin.
func (d ArchiveMixin) Hooks() []ent.Hook {
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

// P adds a storage-level predicate to the queries and mutations.
func (d ArchiveMixin) P(w interface{ WhereP(...func(*sql.Selector)) }) {
	w.WhereP(
		sql.FieldIsNull(d.Fields()[0].Descriptor().Name),
	)
}
