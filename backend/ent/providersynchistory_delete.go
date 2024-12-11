// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/providersynchistory"
)

// ProviderSyncHistoryDelete is the builder for deleting a ProviderSyncHistory entity.
type ProviderSyncHistoryDelete struct {
	config
	hooks    []Hook
	mutation *ProviderSyncHistoryMutation
}

// Where appends a list predicates to the ProviderSyncHistoryDelete builder.
func (pshd *ProviderSyncHistoryDelete) Where(ps ...predicate.ProviderSyncHistory) *ProviderSyncHistoryDelete {
	pshd.mutation.Where(ps...)
	return pshd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (pshd *ProviderSyncHistoryDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, pshd.sqlExec, pshd.mutation, pshd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (pshd *ProviderSyncHistoryDelete) ExecX(ctx context.Context) int {
	n, err := pshd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (pshd *ProviderSyncHistoryDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(providersynchistory.Table, sqlgraph.NewFieldSpec(providersynchistory.FieldID, field.TypeUUID))
	if ps := pshd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, pshd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	pshd.mutation.done = true
	return affected, err
}

// ProviderSyncHistoryDeleteOne is the builder for deleting a single ProviderSyncHistory entity.
type ProviderSyncHistoryDeleteOne struct {
	pshd *ProviderSyncHistoryDelete
}

// Where appends a list predicates to the ProviderSyncHistoryDelete builder.
func (pshdo *ProviderSyncHistoryDeleteOne) Where(ps ...predicate.ProviderSyncHistory) *ProviderSyncHistoryDeleteOne {
	pshdo.pshd.mutation.Where(ps...)
	return pshdo
}

// Exec executes the deletion query.
func (pshdo *ProviderSyncHistoryDeleteOne) Exec(ctx context.Context) error {
	n, err := pshdo.pshd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{providersynchistory.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (pshdo *ProviderSyncHistoryDeleteOne) ExecX(ctx context.Context) {
	if err := pshdo.Exec(ctx); err != nil {
		panic(err)
	}
}
