// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/twohundreds/rezible/ent/predicate"
	"github.com/twohundreds/rezible/ent/providerconfig"
)

// ProviderConfigDelete is the builder for deleting a ProviderConfig entity.
type ProviderConfigDelete struct {
	config
	hooks    []Hook
	mutation *ProviderConfigMutation
}

// Where appends a list predicates to the ProviderConfigDelete builder.
func (pcd *ProviderConfigDelete) Where(ps ...predicate.ProviderConfig) *ProviderConfigDelete {
	pcd.mutation.Where(ps...)
	return pcd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (pcd *ProviderConfigDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, pcd.sqlExec, pcd.mutation, pcd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (pcd *ProviderConfigDelete) ExecX(ctx context.Context) int {
	n, err := pcd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (pcd *ProviderConfigDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(providerconfig.Table, sqlgraph.NewFieldSpec(providerconfig.FieldID, field.TypeUUID))
	if ps := pcd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, pcd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	pcd.mutation.done = true
	return affected, err
}

// ProviderConfigDeleteOne is the builder for deleting a single ProviderConfig entity.
type ProviderConfigDeleteOne struct {
	pcd *ProviderConfigDelete
}

// Where appends a list predicates to the ProviderConfigDelete builder.
func (pcdo *ProviderConfigDeleteOne) Where(ps ...predicate.ProviderConfig) *ProviderConfigDeleteOne {
	pcdo.pcd.mutation.Where(ps...)
	return pcdo
}

// Exec executes the deletion query.
func (pcdo *ProviderConfigDeleteOne) Exec(ctx context.Context) error {
	n, err := pcdo.pcd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{providerconfig.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (pcdo *ProviderConfigDeleteOne) ExecX(ctx context.Context) {
	if err := pcdo.Exec(ctx); err != nil {
		panic(err)
	}
}