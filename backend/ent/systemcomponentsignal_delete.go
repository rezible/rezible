// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/systemcomponentsignal"
)

// SystemComponentSignalDelete is the builder for deleting a SystemComponentSignal entity.
type SystemComponentSignalDelete struct {
	config
	hooks    []Hook
	mutation *SystemComponentSignalMutation
}

// Where appends a list predicates to the SystemComponentSignalDelete builder.
func (scsd *SystemComponentSignalDelete) Where(ps ...predicate.SystemComponentSignal) *SystemComponentSignalDelete {
	scsd.mutation.Where(ps...)
	return scsd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (scsd *SystemComponentSignalDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, scsd.sqlExec, scsd.mutation, scsd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (scsd *SystemComponentSignalDelete) ExecX(ctx context.Context) int {
	n, err := scsd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (scsd *SystemComponentSignalDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(systemcomponentsignal.Table, sqlgraph.NewFieldSpec(systemcomponentsignal.FieldID, field.TypeUUID))
	if ps := scsd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, scsd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	scsd.mutation.done = true
	return affected, err
}

// SystemComponentSignalDeleteOne is the builder for deleting a single SystemComponentSignal entity.
type SystemComponentSignalDeleteOne struct {
	scsd *SystemComponentSignalDelete
}

// Where appends a list predicates to the SystemComponentSignalDelete builder.
func (scsdo *SystemComponentSignalDeleteOne) Where(ps ...predicate.SystemComponentSignal) *SystemComponentSignalDeleteOne {
	scsdo.scsd.mutation.Where(ps...)
	return scsdo
}

// Exec executes the deletion query.
func (scsdo *SystemComponentSignalDeleteOne) Exec(ctx context.Context) error {
	n, err := scsdo.scsd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{systemcomponentsignal.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (scsdo *SystemComponentSignalDeleteOne) ExecX(ctx context.Context) {
	if err := scsdo.Exec(ctx); err != nil {
		panic(err)
	}
}
