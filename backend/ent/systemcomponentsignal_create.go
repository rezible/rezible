// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/systemcomponent"
	"github.com/rezible/rezible/ent/systemcomponentrelationshipfeedback"
	"github.com/rezible/rezible/ent/systemcomponentsignal"
)

// SystemComponentSignalCreate is the builder for creating a SystemComponentSignal entity.
type SystemComponentSignalCreate struct {
	config
	mutation *SystemComponentSignalMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetComponentID sets the "component_id" field.
func (scsc *SystemComponentSignalCreate) SetComponentID(u uuid.UUID) *SystemComponentSignalCreate {
	scsc.mutation.SetComponentID(u)
	return scsc
}

// SetDescription sets the "description" field.
func (scsc *SystemComponentSignalCreate) SetDescription(s string) *SystemComponentSignalCreate {
	scsc.mutation.SetDescription(s)
	return scsc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (scsc *SystemComponentSignalCreate) SetNillableDescription(s *string) *SystemComponentSignalCreate {
	if s != nil {
		scsc.SetDescription(*s)
	}
	return scsc
}

// SetCreatedAt sets the "created_at" field.
func (scsc *SystemComponentSignalCreate) SetCreatedAt(t time.Time) *SystemComponentSignalCreate {
	scsc.mutation.SetCreatedAt(t)
	return scsc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (scsc *SystemComponentSignalCreate) SetNillableCreatedAt(t *time.Time) *SystemComponentSignalCreate {
	if t != nil {
		scsc.SetCreatedAt(*t)
	}
	return scsc
}

// SetID sets the "id" field.
func (scsc *SystemComponentSignalCreate) SetID(u uuid.UUID) *SystemComponentSignalCreate {
	scsc.mutation.SetID(u)
	return scsc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (scsc *SystemComponentSignalCreate) SetNillableID(u *uuid.UUID) *SystemComponentSignalCreate {
	if u != nil {
		scsc.SetID(*u)
	}
	return scsc
}

// SetComponent sets the "component" edge to the SystemComponent entity.
func (scsc *SystemComponentSignalCreate) SetComponent(s *SystemComponent) *SystemComponentSignalCreate {
	return scsc.SetComponentID(s.ID)
}

// AddFeedbackSignalIDs adds the "feedback_signals" edge to the SystemComponentRelationshipFeedback entity by IDs.
func (scsc *SystemComponentSignalCreate) AddFeedbackSignalIDs(ids ...uuid.UUID) *SystemComponentSignalCreate {
	scsc.mutation.AddFeedbackSignalIDs(ids...)
	return scsc
}

// AddFeedbackSignals adds the "feedback_signals" edges to the SystemComponentRelationshipFeedback entity.
func (scsc *SystemComponentSignalCreate) AddFeedbackSignals(s ...*SystemComponentRelationshipFeedback) *SystemComponentSignalCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return scsc.AddFeedbackSignalIDs(ids...)
}

// Mutation returns the SystemComponentSignalMutation object of the builder.
func (scsc *SystemComponentSignalCreate) Mutation() *SystemComponentSignalMutation {
	return scsc.mutation
}

// Save creates the SystemComponentSignal in the database.
func (scsc *SystemComponentSignalCreate) Save(ctx context.Context) (*SystemComponentSignal, error) {
	scsc.defaults()
	return withHooks(ctx, scsc.sqlSave, scsc.mutation, scsc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (scsc *SystemComponentSignalCreate) SaveX(ctx context.Context) *SystemComponentSignal {
	v, err := scsc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (scsc *SystemComponentSignalCreate) Exec(ctx context.Context) error {
	_, err := scsc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (scsc *SystemComponentSignalCreate) ExecX(ctx context.Context) {
	if err := scsc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (scsc *SystemComponentSignalCreate) defaults() {
	if _, ok := scsc.mutation.CreatedAt(); !ok {
		v := systemcomponentsignal.DefaultCreatedAt()
		scsc.mutation.SetCreatedAt(v)
	}
	if _, ok := scsc.mutation.ID(); !ok {
		v := systemcomponentsignal.DefaultID()
		scsc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (scsc *SystemComponentSignalCreate) check() error {
	if _, ok := scsc.mutation.ComponentID(); !ok {
		return &ValidationError{Name: "component_id", err: errors.New(`ent: missing required field "SystemComponentSignal.component_id"`)}
	}
	if _, ok := scsc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "SystemComponentSignal.created_at"`)}
	}
	if len(scsc.mutation.ComponentIDs()) == 0 {
		return &ValidationError{Name: "component", err: errors.New(`ent: missing required edge "SystemComponentSignal.component"`)}
	}
	return nil
}

func (scsc *SystemComponentSignalCreate) sqlSave(ctx context.Context) (*SystemComponentSignal, error) {
	if err := scsc.check(); err != nil {
		return nil, err
	}
	_node, _spec := scsc.createSpec()
	if err := sqlgraph.CreateNode(ctx, scsc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(*uuid.UUID); ok {
			_node.ID = *id
		} else if err := _node.ID.Scan(_spec.ID.Value); err != nil {
			return nil, err
		}
	}
	scsc.mutation.id = &_node.ID
	scsc.mutation.done = true
	return _node, nil
}

func (scsc *SystemComponentSignalCreate) createSpec() (*SystemComponentSignal, *sqlgraph.CreateSpec) {
	var (
		_node = &SystemComponentSignal{config: scsc.config}
		_spec = sqlgraph.NewCreateSpec(systemcomponentsignal.Table, sqlgraph.NewFieldSpec(systemcomponentsignal.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = scsc.conflict
	if id, ok := scsc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := scsc.mutation.Description(); ok {
		_spec.SetField(systemcomponentsignal.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if value, ok := scsc.mutation.CreatedAt(); ok {
		_spec.SetField(systemcomponentsignal.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if nodes := scsc.mutation.ComponentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentsignal.ComponentTable,
			Columns: []string{systemcomponentsignal.ComponentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.ComponentID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := scsc.mutation.FeedbackSignalsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemcomponentsignal.FeedbackSignalsTable,
			Columns: []string{systemcomponentsignal.FeedbackSignalsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentrelationshipfeedback.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.SystemComponentSignal.Create().
//		SetComponentID(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.SystemComponentSignalUpsert) {
//			SetComponentID(v+v).
//		}).
//		Exec(ctx)
func (scsc *SystemComponentSignalCreate) OnConflict(opts ...sql.ConflictOption) *SystemComponentSignalUpsertOne {
	scsc.conflict = opts
	return &SystemComponentSignalUpsertOne{
		create: scsc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.SystemComponentSignal.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (scsc *SystemComponentSignalCreate) OnConflictColumns(columns ...string) *SystemComponentSignalUpsertOne {
	scsc.conflict = append(scsc.conflict, sql.ConflictColumns(columns...))
	return &SystemComponentSignalUpsertOne{
		create: scsc,
	}
}

type (
	// SystemComponentSignalUpsertOne is the builder for "upsert"-ing
	//  one SystemComponentSignal node.
	SystemComponentSignalUpsertOne struct {
		create *SystemComponentSignalCreate
	}

	// SystemComponentSignalUpsert is the "OnConflict" setter.
	SystemComponentSignalUpsert struct {
		*sql.UpdateSet
	}
)

// SetComponentID sets the "component_id" field.
func (u *SystemComponentSignalUpsert) SetComponentID(v uuid.UUID) *SystemComponentSignalUpsert {
	u.Set(systemcomponentsignal.FieldComponentID, v)
	return u
}

// UpdateComponentID sets the "component_id" field to the value that was provided on create.
func (u *SystemComponentSignalUpsert) UpdateComponentID() *SystemComponentSignalUpsert {
	u.SetExcluded(systemcomponentsignal.FieldComponentID)
	return u
}

// SetDescription sets the "description" field.
func (u *SystemComponentSignalUpsert) SetDescription(v string) *SystemComponentSignalUpsert {
	u.Set(systemcomponentsignal.FieldDescription, v)
	return u
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *SystemComponentSignalUpsert) UpdateDescription() *SystemComponentSignalUpsert {
	u.SetExcluded(systemcomponentsignal.FieldDescription)
	return u
}

// ClearDescription clears the value of the "description" field.
func (u *SystemComponentSignalUpsert) ClearDescription() *SystemComponentSignalUpsert {
	u.SetNull(systemcomponentsignal.FieldDescription)
	return u
}

// SetCreatedAt sets the "created_at" field.
func (u *SystemComponentSignalUpsert) SetCreatedAt(v time.Time) *SystemComponentSignalUpsert {
	u.Set(systemcomponentsignal.FieldCreatedAt, v)
	return u
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *SystemComponentSignalUpsert) UpdateCreatedAt() *SystemComponentSignalUpsert {
	u.SetExcluded(systemcomponentsignal.FieldCreatedAt)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.SystemComponentSignal.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(systemcomponentsignal.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *SystemComponentSignalUpsertOne) UpdateNewValues() *SystemComponentSignalUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(systemcomponentsignal.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.SystemComponentSignal.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *SystemComponentSignalUpsertOne) Ignore() *SystemComponentSignalUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *SystemComponentSignalUpsertOne) DoNothing() *SystemComponentSignalUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the SystemComponentSignalCreate.OnConflict
// documentation for more info.
func (u *SystemComponentSignalUpsertOne) Update(set func(*SystemComponentSignalUpsert)) *SystemComponentSignalUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&SystemComponentSignalUpsert{UpdateSet: update})
	}))
	return u
}

// SetComponentID sets the "component_id" field.
func (u *SystemComponentSignalUpsertOne) SetComponentID(v uuid.UUID) *SystemComponentSignalUpsertOne {
	return u.Update(func(s *SystemComponentSignalUpsert) {
		s.SetComponentID(v)
	})
}

// UpdateComponentID sets the "component_id" field to the value that was provided on create.
func (u *SystemComponentSignalUpsertOne) UpdateComponentID() *SystemComponentSignalUpsertOne {
	return u.Update(func(s *SystemComponentSignalUpsert) {
		s.UpdateComponentID()
	})
}

// SetDescription sets the "description" field.
func (u *SystemComponentSignalUpsertOne) SetDescription(v string) *SystemComponentSignalUpsertOne {
	return u.Update(func(s *SystemComponentSignalUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *SystemComponentSignalUpsertOne) UpdateDescription() *SystemComponentSignalUpsertOne {
	return u.Update(func(s *SystemComponentSignalUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *SystemComponentSignalUpsertOne) ClearDescription() *SystemComponentSignalUpsertOne {
	return u.Update(func(s *SystemComponentSignalUpsert) {
		s.ClearDescription()
	})
}

// SetCreatedAt sets the "created_at" field.
func (u *SystemComponentSignalUpsertOne) SetCreatedAt(v time.Time) *SystemComponentSignalUpsertOne {
	return u.Update(func(s *SystemComponentSignalUpsert) {
		s.SetCreatedAt(v)
	})
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *SystemComponentSignalUpsertOne) UpdateCreatedAt() *SystemComponentSignalUpsertOne {
	return u.Update(func(s *SystemComponentSignalUpsert) {
		s.UpdateCreatedAt()
	})
}

// Exec executes the query.
func (u *SystemComponentSignalUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for SystemComponentSignalCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *SystemComponentSignalUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *SystemComponentSignalUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: SystemComponentSignalUpsertOne.ID is not supported by MySQL driver. Use SystemComponentSignalUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *SystemComponentSignalUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// SystemComponentSignalCreateBulk is the builder for creating many SystemComponentSignal entities in bulk.
type SystemComponentSignalCreateBulk struct {
	config
	err      error
	builders []*SystemComponentSignalCreate
	conflict []sql.ConflictOption
}

// Save creates the SystemComponentSignal entities in the database.
func (scscb *SystemComponentSignalCreateBulk) Save(ctx context.Context) ([]*SystemComponentSignal, error) {
	if scscb.err != nil {
		return nil, scscb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(scscb.builders))
	nodes := make([]*SystemComponentSignal, len(scscb.builders))
	mutators := make([]Mutator, len(scscb.builders))
	for i := range scscb.builders {
		func(i int, root context.Context) {
			builder := scscb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SystemComponentSignalMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, scscb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = scscb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, scscb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, scscb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (scscb *SystemComponentSignalCreateBulk) SaveX(ctx context.Context) []*SystemComponentSignal {
	v, err := scscb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (scscb *SystemComponentSignalCreateBulk) Exec(ctx context.Context) error {
	_, err := scscb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (scscb *SystemComponentSignalCreateBulk) ExecX(ctx context.Context) {
	if err := scscb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.SystemComponentSignal.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.SystemComponentSignalUpsert) {
//			SetComponentID(v+v).
//		}).
//		Exec(ctx)
func (scscb *SystemComponentSignalCreateBulk) OnConflict(opts ...sql.ConflictOption) *SystemComponentSignalUpsertBulk {
	scscb.conflict = opts
	return &SystemComponentSignalUpsertBulk{
		create: scscb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.SystemComponentSignal.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (scscb *SystemComponentSignalCreateBulk) OnConflictColumns(columns ...string) *SystemComponentSignalUpsertBulk {
	scscb.conflict = append(scscb.conflict, sql.ConflictColumns(columns...))
	return &SystemComponentSignalUpsertBulk{
		create: scscb,
	}
}

// SystemComponentSignalUpsertBulk is the builder for "upsert"-ing
// a bulk of SystemComponentSignal nodes.
type SystemComponentSignalUpsertBulk struct {
	create *SystemComponentSignalCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.SystemComponentSignal.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(systemcomponentsignal.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *SystemComponentSignalUpsertBulk) UpdateNewValues() *SystemComponentSignalUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(systemcomponentsignal.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.SystemComponentSignal.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *SystemComponentSignalUpsertBulk) Ignore() *SystemComponentSignalUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *SystemComponentSignalUpsertBulk) DoNothing() *SystemComponentSignalUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the SystemComponentSignalCreateBulk.OnConflict
// documentation for more info.
func (u *SystemComponentSignalUpsertBulk) Update(set func(*SystemComponentSignalUpsert)) *SystemComponentSignalUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&SystemComponentSignalUpsert{UpdateSet: update})
	}))
	return u
}

// SetComponentID sets the "component_id" field.
func (u *SystemComponentSignalUpsertBulk) SetComponentID(v uuid.UUID) *SystemComponentSignalUpsertBulk {
	return u.Update(func(s *SystemComponentSignalUpsert) {
		s.SetComponentID(v)
	})
}

// UpdateComponentID sets the "component_id" field to the value that was provided on create.
func (u *SystemComponentSignalUpsertBulk) UpdateComponentID() *SystemComponentSignalUpsertBulk {
	return u.Update(func(s *SystemComponentSignalUpsert) {
		s.UpdateComponentID()
	})
}

// SetDescription sets the "description" field.
func (u *SystemComponentSignalUpsertBulk) SetDescription(v string) *SystemComponentSignalUpsertBulk {
	return u.Update(func(s *SystemComponentSignalUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *SystemComponentSignalUpsertBulk) UpdateDescription() *SystemComponentSignalUpsertBulk {
	return u.Update(func(s *SystemComponentSignalUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *SystemComponentSignalUpsertBulk) ClearDescription() *SystemComponentSignalUpsertBulk {
	return u.Update(func(s *SystemComponentSignalUpsert) {
		s.ClearDescription()
	})
}

// SetCreatedAt sets the "created_at" field.
func (u *SystemComponentSignalUpsertBulk) SetCreatedAt(v time.Time) *SystemComponentSignalUpsertBulk {
	return u.Update(func(s *SystemComponentSignalUpsert) {
		s.SetCreatedAt(v)
	})
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *SystemComponentSignalUpsertBulk) UpdateCreatedAt() *SystemComponentSignalUpsertBulk {
	return u.Update(func(s *SystemComponentSignalUpsert) {
		s.UpdateCreatedAt()
	})
}

// Exec executes the query.
func (u *SystemComponentSignalUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the SystemComponentSignalCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for SystemComponentSignalCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *SystemComponentSignalUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
