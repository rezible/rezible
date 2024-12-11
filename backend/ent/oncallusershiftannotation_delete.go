// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/rezible/rezible/ent/oncallusershiftannotation"
	"github.com/rezible/rezible/ent/predicate"
)

// OncallUserShiftAnnotationDelete is the builder for deleting a OncallUserShiftAnnotation entity.
type OncallUserShiftAnnotationDelete struct {
	config
	hooks    []Hook
	mutation *OncallUserShiftAnnotationMutation
}

// Where appends a list predicates to the OncallUserShiftAnnotationDelete builder.
func (ousad *OncallUserShiftAnnotationDelete) Where(ps ...predicate.OncallUserShiftAnnotation) *OncallUserShiftAnnotationDelete {
	ousad.mutation.Where(ps...)
	return ousad
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (ousad *OncallUserShiftAnnotationDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, ousad.sqlExec, ousad.mutation, ousad.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (ousad *OncallUserShiftAnnotationDelete) ExecX(ctx context.Context) int {
	n, err := ousad.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (ousad *OncallUserShiftAnnotationDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(oncallusershiftannotation.Table, sqlgraph.NewFieldSpec(oncallusershiftannotation.FieldID, field.TypeUUID))
	if ps := ousad.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, ousad.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	ousad.mutation.done = true
	return affected, err
}

// OncallUserShiftAnnotationDeleteOne is the builder for deleting a single OncallUserShiftAnnotation entity.
type OncallUserShiftAnnotationDeleteOne struct {
	ousad *OncallUserShiftAnnotationDelete
}

// Where appends a list predicates to the OncallUserShiftAnnotationDelete builder.
func (ousado *OncallUserShiftAnnotationDeleteOne) Where(ps ...predicate.OncallUserShiftAnnotation) *OncallUserShiftAnnotationDeleteOne {
	ousado.ousad.mutation.Where(ps...)
	return ousado
}

// Exec executes the deletion query.
func (ousado *OncallUserShiftAnnotationDeleteOne) Exec(ctx context.Context) error {
	n, err := ousado.ousad.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{oncallusershiftannotation.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (ousado *OncallUserShiftAnnotationDeleteOne) ExecX(ctx context.Context) {
	if err := ousado.Exec(ctx); err != nil {
		panic(err)
	}
}
