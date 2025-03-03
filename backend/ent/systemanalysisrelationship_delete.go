// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/systemanalysisrelationship"
)

// SystemAnalysisRelationshipDelete is the builder for deleting a SystemAnalysisRelationship entity.
type SystemAnalysisRelationshipDelete struct {
	config
	hooks    []Hook
	mutation *SystemAnalysisRelationshipMutation
}

// Where appends a list predicates to the SystemAnalysisRelationshipDelete builder.
func (sard *SystemAnalysisRelationshipDelete) Where(ps ...predicate.SystemAnalysisRelationship) *SystemAnalysisRelationshipDelete {
	sard.mutation.Where(ps...)
	return sard
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (sard *SystemAnalysisRelationshipDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, sard.sqlExec, sard.mutation, sard.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (sard *SystemAnalysisRelationshipDelete) ExecX(ctx context.Context) int {
	n, err := sard.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (sard *SystemAnalysisRelationshipDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(systemanalysisrelationship.Table, sqlgraph.NewFieldSpec(systemanalysisrelationship.FieldID, field.TypeUUID))
	if ps := sard.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, sard.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	sard.mutation.done = true
	return affected, err
}

// SystemAnalysisRelationshipDeleteOne is the builder for deleting a single SystemAnalysisRelationship entity.
type SystemAnalysisRelationshipDeleteOne struct {
	sard *SystemAnalysisRelationshipDelete
}

// Where appends a list predicates to the SystemAnalysisRelationshipDelete builder.
func (sardo *SystemAnalysisRelationshipDeleteOne) Where(ps ...predicate.SystemAnalysisRelationship) *SystemAnalysisRelationshipDeleteOne {
	sardo.sard.mutation.Where(ps...)
	return sardo
}

// Exec executes the deletion query.
func (sardo *SystemAnalysisRelationshipDeleteOne) Exec(ctx context.Context) error {
	n, err := sardo.sard.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{systemanalysisrelationship.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (sardo *SystemAnalysisRelationshipDeleteOne) ExecX(ctx context.Context) {
	if err := sardo.Exec(ctx); err != nil {
		panic(err)
	}
}
