// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/rezible/rezible/ent/oncallusershiftcover"
	"github.com/rezible/rezible/ent/predicate"
)

// OncallUserShiftCoverDelete is the builder for deleting a OncallUserShiftCover entity.
type OncallUserShiftCoverDelete struct {
	config
	hooks    []Hook
	mutation *OncallUserShiftCoverMutation
}

// Where appends a list predicates to the OncallUserShiftCoverDelete builder.
func (ouscd *OncallUserShiftCoverDelete) Where(ps ...predicate.OncallUserShiftCover) *OncallUserShiftCoverDelete {
	ouscd.mutation.Where(ps...)
	return ouscd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (ouscd *OncallUserShiftCoverDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, ouscd.sqlExec, ouscd.mutation, ouscd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (ouscd *OncallUserShiftCoverDelete) ExecX(ctx context.Context) int {
	n, err := ouscd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (ouscd *OncallUserShiftCoverDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(oncallusershiftcover.Table, sqlgraph.NewFieldSpec(oncallusershiftcover.FieldID, field.TypeUUID))
	if ps := ouscd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, ouscd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	ouscd.mutation.done = true
	return affected, err
}

// OncallUserShiftCoverDeleteOne is the builder for deleting a single OncallUserShiftCover entity.
type OncallUserShiftCoverDeleteOne struct {
	ouscd *OncallUserShiftCoverDelete
}

// Where appends a list predicates to the OncallUserShiftCoverDelete builder.
func (ouscdo *OncallUserShiftCoverDeleteOne) Where(ps ...predicate.OncallUserShiftCover) *OncallUserShiftCoverDeleteOne {
	ouscdo.ouscd.mutation.Where(ps...)
	return ouscdo
}

// Exec executes the deletion query.
func (ouscdo *OncallUserShiftCoverDeleteOne) Exec(ctx context.Context) error {
	n, err := ouscdo.ouscd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{oncallusershiftcover.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (ouscdo *OncallUserShiftCoverDeleteOne) ExecX(ctx context.Context) {
	if err := ouscdo.Exec(ctx); err != nil {
		panic(err)
	}
}
