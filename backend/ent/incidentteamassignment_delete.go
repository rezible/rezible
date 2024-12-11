// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/rezible/rezible/ent/incidentteamassignment"
	"github.com/rezible/rezible/ent/predicate"
)

// IncidentTeamAssignmentDelete is the builder for deleting a IncidentTeamAssignment entity.
type IncidentTeamAssignmentDelete struct {
	config
	hooks    []Hook
	mutation *IncidentTeamAssignmentMutation
}

// Where appends a list predicates to the IncidentTeamAssignmentDelete builder.
func (itad *IncidentTeamAssignmentDelete) Where(ps ...predicate.IncidentTeamAssignment) *IncidentTeamAssignmentDelete {
	itad.mutation.Where(ps...)
	return itad
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (itad *IncidentTeamAssignmentDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, itad.sqlExec, itad.mutation, itad.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (itad *IncidentTeamAssignmentDelete) ExecX(ctx context.Context) int {
	n, err := itad.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (itad *IncidentTeamAssignmentDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(incidentteamassignment.Table, sqlgraph.NewFieldSpec(incidentteamassignment.FieldID, field.TypeInt))
	if ps := itad.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, itad.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	itad.mutation.done = true
	return affected, err
}

// IncidentTeamAssignmentDeleteOne is the builder for deleting a single IncidentTeamAssignment entity.
type IncidentTeamAssignmentDeleteOne struct {
	itad *IncidentTeamAssignmentDelete
}

// Where appends a list predicates to the IncidentTeamAssignmentDelete builder.
func (itado *IncidentTeamAssignmentDeleteOne) Where(ps ...predicate.IncidentTeamAssignment) *IncidentTeamAssignmentDeleteOne {
	itado.itad.mutation.Where(ps...)
	return itado
}

// Exec executes the deletion query.
func (itado *IncidentTeamAssignmentDeleteOne) Exec(ctx context.Context) error {
	n, err := itado.itad.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{incidentteamassignment.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (itado *IncidentTeamAssignmentDeleteOne) ExecX(ctx context.Context) {
	if err := itado.Exec(ctx); err != nil {
		panic(err)
	}
}
