// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/twohundreds/rezible/ent/predicate"
	"github.com/twohundreds/rezible/ent/retrospectivereview"
)

// RetrospectiveReviewDelete is the builder for deleting a RetrospectiveReview entity.
type RetrospectiveReviewDelete struct {
	config
	hooks    []Hook
	mutation *RetrospectiveReviewMutation
}

// Where appends a list predicates to the RetrospectiveReviewDelete builder.
func (rrd *RetrospectiveReviewDelete) Where(ps ...predicate.RetrospectiveReview) *RetrospectiveReviewDelete {
	rrd.mutation.Where(ps...)
	return rrd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (rrd *RetrospectiveReviewDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, rrd.sqlExec, rrd.mutation, rrd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (rrd *RetrospectiveReviewDelete) ExecX(ctx context.Context) int {
	n, err := rrd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (rrd *RetrospectiveReviewDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(retrospectivereview.Table, sqlgraph.NewFieldSpec(retrospectivereview.FieldID, field.TypeUUID))
	if ps := rrd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, rrd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	rrd.mutation.done = true
	return affected, err
}

// RetrospectiveReviewDeleteOne is the builder for deleting a single RetrospectiveReview entity.
type RetrospectiveReviewDeleteOne struct {
	rrd *RetrospectiveReviewDelete
}

// Where appends a list predicates to the RetrospectiveReviewDelete builder.
func (rrdo *RetrospectiveReviewDeleteOne) Where(ps ...predicate.RetrospectiveReview) *RetrospectiveReviewDeleteOne {
	rrdo.rrd.mutation.Where(ps...)
	return rrdo
}

// Exec executes the deletion query.
func (rrdo *RetrospectiveReviewDeleteOne) Exec(ctx context.Context) error {
	n, err := rrdo.rrd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{retrospectivereview.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (rrdo *RetrospectiveReviewDeleteOne) ExecX(ctx context.Context) {
	if err := rrdo.Exec(ctx); err != nil {
		panic(err)
	}
}