// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/twohundreds/rezible/ent/incidentdebriefmessage"
	"github.com/twohundreds/rezible/ent/predicate"
)

// IncidentDebriefMessageDelete is the builder for deleting a IncidentDebriefMessage entity.
type IncidentDebriefMessageDelete struct {
	config
	hooks    []Hook
	mutation *IncidentDebriefMessageMutation
}

// Where appends a list predicates to the IncidentDebriefMessageDelete builder.
func (idmd *IncidentDebriefMessageDelete) Where(ps ...predicate.IncidentDebriefMessage) *IncidentDebriefMessageDelete {
	idmd.mutation.Where(ps...)
	return idmd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (idmd *IncidentDebriefMessageDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, idmd.sqlExec, idmd.mutation, idmd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (idmd *IncidentDebriefMessageDelete) ExecX(ctx context.Context) int {
	n, err := idmd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (idmd *IncidentDebriefMessageDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(incidentdebriefmessage.Table, sqlgraph.NewFieldSpec(incidentdebriefmessage.FieldID, field.TypeUUID))
	if ps := idmd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, idmd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	idmd.mutation.done = true
	return affected, err
}

// IncidentDebriefMessageDeleteOne is the builder for deleting a single IncidentDebriefMessage entity.
type IncidentDebriefMessageDeleteOne struct {
	idmd *IncidentDebriefMessageDelete
}

// Where appends a list predicates to the IncidentDebriefMessageDelete builder.
func (idmdo *IncidentDebriefMessageDeleteOne) Where(ps ...predicate.IncidentDebriefMessage) *IncidentDebriefMessageDeleteOne {
	idmdo.idmd.mutation.Where(ps...)
	return idmdo
}

// Exec executes the deletion query.
func (idmdo *IncidentDebriefMessageDeleteOne) Exec(ctx context.Context) error {
	n, err := idmdo.idmd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{incidentdebriefmessage.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (idmdo *IncidentDebriefMessageDeleteOne) ExecX(ctx context.Context) {
	if err := idmdo.Exec(ctx); err != nil {
		panic(err)
	}
}