// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/rezible/rezible/ent/incidenteventcontext"
	"github.com/rezible/rezible/ent/predicate"
)

// IncidentEventContextDelete is the builder for deleting a IncidentEventContext entity.
type IncidentEventContextDelete struct {
	config
	hooks    []Hook
	mutation *IncidentEventContextMutation
}

// Where appends a list predicates to the IncidentEventContextDelete builder.
func (iecd *IncidentEventContextDelete) Where(ps ...predicate.IncidentEventContext) *IncidentEventContextDelete {
	iecd.mutation.Where(ps...)
	return iecd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (iecd *IncidentEventContextDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, iecd.sqlExec, iecd.mutation, iecd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (iecd *IncidentEventContextDelete) ExecX(ctx context.Context) int {
	n, err := iecd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (iecd *IncidentEventContextDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(incidenteventcontext.Table, sqlgraph.NewFieldSpec(incidenteventcontext.FieldID, field.TypeUUID))
	if ps := iecd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, iecd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	iecd.mutation.done = true
	return affected, err
}

// IncidentEventContextDeleteOne is the builder for deleting a single IncidentEventContext entity.
type IncidentEventContextDeleteOne struct {
	iecd *IncidentEventContextDelete
}

// Where appends a list predicates to the IncidentEventContextDelete builder.
func (iecdo *IncidentEventContextDeleteOne) Where(ps ...predicate.IncidentEventContext) *IncidentEventContextDeleteOne {
	iecdo.iecd.mutation.Where(ps...)
	return iecdo
}

// Exec executes the deletion query.
func (iecdo *IncidentEventContextDeleteOne) Exec(ctx context.Context) error {
	n, err := iecdo.iecd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{incidenteventcontext.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (iecdo *IncidentEventContextDeleteOne) ExecX(ctx context.Context) {
	if err := iecdo.Exec(ctx); err != nil {
		panic(err)
	}
}
