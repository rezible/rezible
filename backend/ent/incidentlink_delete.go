// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/twohundreds/rezible/ent/incidentlink"
	"github.com/twohundreds/rezible/ent/predicate"
)

// IncidentLinkDelete is the builder for deleting a IncidentLink entity.
type IncidentLinkDelete struct {
	config
	hooks    []Hook
	mutation *IncidentLinkMutation
}

// Where appends a list predicates to the IncidentLinkDelete builder.
func (ild *IncidentLinkDelete) Where(ps ...predicate.IncidentLink) *IncidentLinkDelete {
	ild.mutation.Where(ps...)
	return ild
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (ild *IncidentLinkDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, ild.sqlExec, ild.mutation, ild.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (ild *IncidentLinkDelete) ExecX(ctx context.Context) int {
	n, err := ild.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (ild *IncidentLinkDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(incidentlink.Table, sqlgraph.NewFieldSpec(incidentlink.FieldID, field.TypeInt))
	if ps := ild.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, ild.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	ild.mutation.done = true
	return affected, err
}

// IncidentLinkDeleteOne is the builder for deleting a single IncidentLink entity.
type IncidentLinkDeleteOne struct {
	ild *IncidentLinkDelete
}

// Where appends a list predicates to the IncidentLinkDelete builder.
func (ildo *IncidentLinkDeleteOne) Where(ps ...predicate.IncidentLink) *IncidentLinkDeleteOne {
	ildo.ild.mutation.Where(ps...)
	return ildo
}

// Exec executes the deletion query.
func (ildo *IncidentLinkDeleteOne) Exec(ctx context.Context) error {
	n, err := ildo.ild.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{incidentlink.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (ildo *IncidentLinkDeleteOne) ExecX(ctx context.Context) {
	if err := ildo.Exec(ctx); err != nil {
		panic(err)
	}
}