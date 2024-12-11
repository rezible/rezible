// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/twohundreds/rezible/ent/incidentdebriefsuggestion"
	"github.com/twohundreds/rezible/ent/predicate"
)

// IncidentDebriefSuggestionDelete is the builder for deleting a IncidentDebriefSuggestion entity.
type IncidentDebriefSuggestionDelete struct {
	config
	hooks    []Hook
	mutation *IncidentDebriefSuggestionMutation
}

// Where appends a list predicates to the IncidentDebriefSuggestionDelete builder.
func (idsd *IncidentDebriefSuggestionDelete) Where(ps ...predicate.IncidentDebriefSuggestion) *IncidentDebriefSuggestionDelete {
	idsd.mutation.Where(ps...)
	return idsd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (idsd *IncidentDebriefSuggestionDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, idsd.sqlExec, idsd.mutation, idsd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (idsd *IncidentDebriefSuggestionDelete) ExecX(ctx context.Context) int {
	n, err := idsd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (idsd *IncidentDebriefSuggestionDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(incidentdebriefsuggestion.Table, sqlgraph.NewFieldSpec(incidentdebriefsuggestion.FieldID, field.TypeUUID))
	if ps := idsd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, idsd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	idsd.mutation.done = true
	return affected, err
}

// IncidentDebriefSuggestionDeleteOne is the builder for deleting a single IncidentDebriefSuggestion entity.
type IncidentDebriefSuggestionDeleteOne struct {
	idsd *IncidentDebriefSuggestionDelete
}

// Where appends a list predicates to the IncidentDebriefSuggestionDelete builder.
func (idsdo *IncidentDebriefSuggestionDeleteOne) Where(ps ...predicate.IncidentDebriefSuggestion) *IncidentDebriefSuggestionDeleteOne {
	idsdo.idsd.mutation.Where(ps...)
	return idsdo
}

// Exec executes the deletion query.
func (idsdo *IncidentDebriefSuggestionDeleteOne) Exec(ctx context.Context) error {
	n, err := idsdo.idsd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{incidentdebriefsuggestion.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (idsdo *IncidentDebriefSuggestionDeleteOne) ExecX(ctx context.Context) {
	if err := idsdo.Exec(ctx); err != nil {
		panic(err)
	}
}