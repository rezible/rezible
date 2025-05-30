// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/predicate"
)

// IncidentDelete is the builder for deleting a Incident entity.
type IncidentDelete struct {
	config
	hooks    []Hook
	mutation *IncidentMutation
}

// Where appends a list predicates to the IncidentDelete builder.
func (id *IncidentDelete) Where(ps ...predicate.Incident) *IncidentDelete {
	id.mutation.Where(ps...)
	return id
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (id *IncidentDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, id.sqlExec, id.mutation, id.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (id *IncidentDelete) ExecX(ctx context.Context) int {
	n, err := id.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (id *IncidentDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(incident.Table, sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID))
	if ps := id.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, id.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	id.mutation.done = true
	return affected, err
}

// IncidentDeleteOne is the builder for deleting a single Incident entity.
type IncidentDeleteOne struct {
	id *IncidentDelete
}

// Where appends a list predicates to the IncidentDelete builder.
func (ido *IncidentDeleteOne) Where(ps ...predicate.Incident) *IncidentDeleteOne {
	ido.id.mutation.Where(ps...)
	return ido
}

// Exec executes the deletion query.
func (ido *IncidentDeleteOne) Exec(ctx context.Context) error {
	n, err := ido.id.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{incident.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (ido *IncidentDeleteOne) ExecX(ctx context.Context) {
	if err := ido.Exec(ctx); err != nil {
		panic(err)
	}
}
