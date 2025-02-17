// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/rezible/rezible/ent/incidentfield"
	"github.com/rezible/rezible/ent/predicate"
)

// IncidentFieldDelete is the builder for deleting a IncidentField entity.
type IncidentFieldDelete struct {
	config
	hooks    []Hook
	mutation *IncidentFieldMutation
}

// Where appends a list predicates to the IncidentFieldDelete builder.
func (ifd *IncidentFieldDelete) Where(ps ...predicate.IncidentField) *IncidentFieldDelete {
	ifd.mutation.Where(ps...)
	return ifd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (ifd *IncidentFieldDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, ifd.sqlExec, ifd.mutation, ifd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (ifd *IncidentFieldDelete) ExecX(ctx context.Context) int {
	n, err := ifd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (ifd *IncidentFieldDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(incidentfield.Table, sqlgraph.NewFieldSpec(incidentfield.FieldID, field.TypeUUID))
	if ps := ifd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, ifd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	ifd.mutation.done = true
	return affected, err
}

// IncidentFieldDeleteOne is the builder for deleting a single IncidentField entity.
type IncidentFieldDeleteOne struct {
	ifd *IncidentFieldDelete
}

// Where appends a list predicates to the IncidentFieldDelete builder.
func (ifdo *IncidentFieldDeleteOne) Where(ps ...predicate.IncidentField) *IncidentFieldDeleteOne {
	ifdo.ifd.mutation.Where(ps...)
	return ifdo
}

// Exec executes the deletion query.
func (ifdo *IncidentFieldDeleteOne) Exec(ctx context.Context) error {
	n, err := ifdo.ifd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{incidentfield.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (ifdo *IncidentFieldDeleteOne) ExecX(ctx context.Context) {
	if err := ifdo.Exec(ctx); err != nil {
		panic(err)
	}
}
