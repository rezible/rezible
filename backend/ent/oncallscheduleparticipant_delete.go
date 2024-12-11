// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/twohundreds/rezible/ent/oncallscheduleparticipant"
	"github.com/twohundreds/rezible/ent/predicate"
)

// OncallScheduleParticipantDelete is the builder for deleting a OncallScheduleParticipant entity.
type OncallScheduleParticipantDelete struct {
	config
	hooks    []Hook
	mutation *OncallScheduleParticipantMutation
}

// Where appends a list predicates to the OncallScheduleParticipantDelete builder.
func (ospd *OncallScheduleParticipantDelete) Where(ps ...predicate.OncallScheduleParticipant) *OncallScheduleParticipantDelete {
	ospd.mutation.Where(ps...)
	return ospd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (ospd *OncallScheduleParticipantDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, ospd.sqlExec, ospd.mutation, ospd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (ospd *OncallScheduleParticipantDelete) ExecX(ctx context.Context) int {
	n, err := ospd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (ospd *OncallScheduleParticipantDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(oncallscheduleparticipant.Table, sqlgraph.NewFieldSpec(oncallscheduleparticipant.FieldID, field.TypeUUID))
	if ps := ospd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, ospd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	ospd.mutation.done = true
	return affected, err
}

// OncallScheduleParticipantDeleteOne is the builder for deleting a single OncallScheduleParticipant entity.
type OncallScheduleParticipantDeleteOne struct {
	ospd *OncallScheduleParticipantDelete
}

// Where appends a list predicates to the OncallScheduleParticipantDelete builder.
func (ospdo *OncallScheduleParticipantDeleteOne) Where(ps ...predicate.OncallScheduleParticipant) *OncallScheduleParticipantDeleteOne {
	ospdo.ospd.mutation.Where(ps...)
	return ospdo
}

// Exec executes the deletion query.
func (ospdo *OncallScheduleParticipantDeleteOne) Exec(ctx context.Context) error {
	n, err := ospdo.ospd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{oncallscheduleparticipant.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (ospdo *OncallScheduleParticipantDeleteOne) ExecX(ctx context.Context) {
	if err := ospdo.Exec(ctx); err != nil {
		panic(err)
	}
}
