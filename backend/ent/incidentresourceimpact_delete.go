// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/twohundreds/rezible/ent/incidentresourceimpact"
	"github.com/twohundreds/rezible/ent/predicate"
)

// IncidentResourceImpactDelete is the builder for deleting a IncidentResourceImpact entity.
type IncidentResourceImpactDelete struct {
	config
	hooks    []Hook
	mutation *IncidentResourceImpactMutation
}

// Where appends a list predicates to the IncidentResourceImpactDelete builder.
func (irid *IncidentResourceImpactDelete) Where(ps ...predicate.IncidentResourceImpact) *IncidentResourceImpactDelete {
	irid.mutation.Where(ps...)
	return irid
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (irid *IncidentResourceImpactDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, irid.sqlExec, irid.mutation, irid.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (irid *IncidentResourceImpactDelete) ExecX(ctx context.Context) int {
	n, err := irid.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (irid *IncidentResourceImpactDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(incidentresourceimpact.Table, sqlgraph.NewFieldSpec(incidentresourceimpact.FieldID, field.TypeUUID))
	if ps := irid.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, irid.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	irid.mutation.done = true
	return affected, err
}

// IncidentResourceImpactDeleteOne is the builder for deleting a single IncidentResourceImpact entity.
type IncidentResourceImpactDeleteOne struct {
	irid *IncidentResourceImpactDelete
}

// Where appends a list predicates to the IncidentResourceImpactDelete builder.
func (irido *IncidentResourceImpactDeleteOne) Where(ps ...predicate.IncidentResourceImpact) *IncidentResourceImpactDeleteOne {
	irido.irid.mutation.Where(ps...)
	return irido
}

// Exec executes the deletion query.
func (irido *IncidentResourceImpactDeleteOne) Exec(ctx context.Context) error {
	n, err := irido.irid.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{incidentresourceimpact.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (irido *IncidentResourceImpactDeleteOne) ExecX(ctx context.Context) {
	if err := irido.Exec(ctx); err != nil {
		panic(err)
	}
}