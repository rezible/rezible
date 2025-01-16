// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/systemcomponentrelationshipfeedback"
)

// SystemComponentRelationshipFeedbackDelete is the builder for deleting a SystemComponentRelationshipFeedback entity.
type SystemComponentRelationshipFeedbackDelete struct {
	config
	hooks    []Hook
	mutation *SystemComponentRelationshipFeedbackMutation
}

// Where appends a list predicates to the SystemComponentRelationshipFeedbackDelete builder.
func (scrfd *SystemComponentRelationshipFeedbackDelete) Where(ps ...predicate.SystemComponentRelationshipFeedback) *SystemComponentRelationshipFeedbackDelete {
	scrfd.mutation.Where(ps...)
	return scrfd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (scrfd *SystemComponentRelationshipFeedbackDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, scrfd.sqlExec, scrfd.mutation, scrfd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (scrfd *SystemComponentRelationshipFeedbackDelete) ExecX(ctx context.Context) int {
	n, err := scrfd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (scrfd *SystemComponentRelationshipFeedbackDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(systemcomponentrelationshipfeedback.Table, sqlgraph.NewFieldSpec(systemcomponentrelationshipfeedback.FieldID, field.TypeUUID))
	if ps := scrfd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, scrfd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	scrfd.mutation.done = true
	return affected, err
}

// SystemComponentRelationshipFeedbackDeleteOne is the builder for deleting a single SystemComponentRelationshipFeedback entity.
type SystemComponentRelationshipFeedbackDeleteOne struct {
	scrfd *SystemComponentRelationshipFeedbackDelete
}

// Where appends a list predicates to the SystemComponentRelationshipFeedbackDelete builder.
func (scrfdo *SystemComponentRelationshipFeedbackDeleteOne) Where(ps ...predicate.SystemComponentRelationshipFeedback) *SystemComponentRelationshipFeedbackDeleteOne {
	scrfdo.scrfd.mutation.Where(ps...)
	return scrfdo
}

// Exec executes the deletion query.
func (scrfdo *SystemComponentRelationshipFeedbackDeleteOne) Exec(ctx context.Context) error {
	n, err := scrfdo.scrfd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{systemcomponentrelationshipfeedback.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (scrfdo *SystemComponentRelationshipFeedbackDeleteOne) ExecX(ctx context.Context) {
	if err := scrfdo.Exec(ctx); err != nil {
		panic(err)
	}
}
