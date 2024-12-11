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
	"github.com/twohundreds/rezible/ent/providersynchistory"
)

// ProviderSyncHistoryCreate is the builder for creating a ProviderSyncHistory entity.
type ProviderSyncHistoryCreate struct {
	config
	mutation *ProviderSyncHistoryMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetDataType sets the "data_type" field.
func (pshc *ProviderSyncHistoryCreate) SetDataType(s string) *ProviderSyncHistoryCreate {
	pshc.mutation.SetDataType(s)
	return pshc
}

// SetStartedAt sets the "started_at" field.
func (pshc *ProviderSyncHistoryCreate) SetStartedAt(t time.Time) *ProviderSyncHistoryCreate {
	pshc.mutation.SetStartedAt(t)
	return pshc
}

// SetNillableStartedAt sets the "started_at" field if the given value is not nil.
func (pshc *ProviderSyncHistoryCreate) SetNillableStartedAt(t *time.Time) *ProviderSyncHistoryCreate {
	if t != nil {
		pshc.SetStartedAt(*t)
	}
	return pshc
}

// SetFinishedAt sets the "finished_at" field.
func (pshc *ProviderSyncHistoryCreate) SetFinishedAt(t time.Time) *ProviderSyncHistoryCreate {
	pshc.mutation.SetFinishedAt(t)
	return pshc
}

// SetNillableFinishedAt sets the "finished_at" field if the given value is not nil.
func (pshc *ProviderSyncHistoryCreate) SetNillableFinishedAt(t *time.Time) *ProviderSyncHistoryCreate {
	if t != nil {
		pshc.SetFinishedAt(*t)
	}
	return pshc
}

// SetNumMutations sets the "num_mutations" field.
func (pshc *ProviderSyncHistoryCreate) SetNumMutations(i int) *ProviderSyncHistoryCreate {
	pshc.mutation.SetNumMutations(i)
	return pshc
}

// SetID sets the "id" field.
func (pshc *ProviderSyncHistoryCreate) SetID(u uuid.UUID) *ProviderSyncHistoryCreate {
	pshc.mutation.SetID(u)
	return pshc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (pshc *ProviderSyncHistoryCreate) SetNillableID(u *uuid.UUID) *ProviderSyncHistoryCreate {
	if u != nil {
		pshc.SetID(*u)
	}
	return pshc
}

// Mutation returns the ProviderSyncHistoryMutation object of the builder.
func (pshc *ProviderSyncHistoryCreate) Mutation() *ProviderSyncHistoryMutation {
	return pshc.mutation
}

// Save creates the ProviderSyncHistory in the database.
func (pshc *ProviderSyncHistoryCreate) Save(ctx context.Context) (*ProviderSyncHistory, error) {
	pshc.defaults()
	return withHooks(ctx, pshc.sqlSave, pshc.mutation, pshc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (pshc *ProviderSyncHistoryCreate) SaveX(ctx context.Context) *ProviderSyncHistory {
	v, err := pshc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (pshc *ProviderSyncHistoryCreate) Exec(ctx context.Context) error {
	_, err := pshc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pshc *ProviderSyncHistoryCreate) ExecX(ctx context.Context) {
	if err := pshc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (pshc *ProviderSyncHistoryCreate) defaults() {
	if _, ok := pshc.mutation.StartedAt(); !ok {
		v := providersynchistory.DefaultStartedAt()
		pshc.mutation.SetStartedAt(v)
	}
	if _, ok := pshc.mutation.FinishedAt(); !ok {
		v := providersynchistory.DefaultFinishedAt()
		pshc.mutation.SetFinishedAt(v)
	}
	if _, ok := pshc.mutation.ID(); !ok {
		v := providersynchistory.DefaultID()
		pshc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (pshc *ProviderSyncHistoryCreate) check() error {
	if _, ok := pshc.mutation.DataType(); !ok {
		return &ValidationError{Name: "data_type", err: errors.New(`ent: missing required field "ProviderSyncHistory.data_type"`)}
	}
	if _, ok := pshc.mutation.StartedAt(); !ok {
		return &ValidationError{Name: "started_at", err: errors.New(`ent: missing required field "ProviderSyncHistory.started_at"`)}
	}
	if _, ok := pshc.mutation.FinishedAt(); !ok {
		return &ValidationError{Name: "finished_at", err: errors.New(`ent: missing required field "ProviderSyncHistory.finished_at"`)}
	}
	if _, ok := pshc.mutation.NumMutations(); !ok {
		return &ValidationError{Name: "num_mutations", err: errors.New(`ent: missing required field "ProviderSyncHistory.num_mutations"`)}
	}
	return nil
}

func (pshc *ProviderSyncHistoryCreate) sqlSave(ctx context.Context) (*ProviderSyncHistory, error) {
	if err := pshc.check(); err != nil {
		return nil, err
	}
	_node, _spec := pshc.createSpec()
	if err := sqlgraph.CreateNode(ctx, pshc.driver, _spec); err != nil {
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
	pshc.mutation.id = &_node.ID
	pshc.mutation.done = true
	return _node, nil
}

func (pshc *ProviderSyncHistoryCreate) createSpec() (*ProviderSyncHistory, *sqlgraph.CreateSpec) {
	var (
		_node = &ProviderSyncHistory{config: pshc.config}
		_spec = sqlgraph.NewCreateSpec(providersynchistory.Table, sqlgraph.NewFieldSpec(providersynchistory.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = pshc.conflict
	if id, ok := pshc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := pshc.mutation.DataType(); ok {
		_spec.SetField(providersynchistory.FieldDataType, field.TypeString, value)
		_node.DataType = value
	}
	if value, ok := pshc.mutation.StartedAt(); ok {
		_spec.SetField(providersynchistory.FieldStartedAt, field.TypeTime, value)
		_node.StartedAt = value
	}
	if value, ok := pshc.mutation.FinishedAt(); ok {
		_spec.SetField(providersynchistory.FieldFinishedAt, field.TypeTime, value)
		_node.FinishedAt = value
	}
	if value, ok := pshc.mutation.NumMutations(); ok {
		_spec.SetField(providersynchistory.FieldNumMutations, field.TypeInt, value)
		_node.NumMutations = value
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.ProviderSyncHistory.Create().
//		SetDataType(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.ProviderSyncHistoryUpsert) {
//			SetDataType(v+v).
//		}).
//		Exec(ctx)
func (pshc *ProviderSyncHistoryCreate) OnConflict(opts ...sql.ConflictOption) *ProviderSyncHistoryUpsertOne {
	pshc.conflict = opts
	return &ProviderSyncHistoryUpsertOne{
		create: pshc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.ProviderSyncHistory.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (pshc *ProviderSyncHistoryCreate) OnConflictColumns(columns ...string) *ProviderSyncHistoryUpsertOne {
	pshc.conflict = append(pshc.conflict, sql.ConflictColumns(columns...))
	return &ProviderSyncHistoryUpsertOne{
		create: pshc,
	}
}

type (
	// ProviderSyncHistoryUpsertOne is the builder for "upsert"-ing
	//  one ProviderSyncHistory node.
	ProviderSyncHistoryUpsertOne struct {
		create *ProviderSyncHistoryCreate
	}

	// ProviderSyncHistoryUpsert is the "OnConflict" setter.
	ProviderSyncHistoryUpsert struct {
		*sql.UpdateSet
	}
)

// SetDataType sets the "data_type" field.
func (u *ProviderSyncHistoryUpsert) SetDataType(v string) *ProviderSyncHistoryUpsert {
	u.Set(providersynchistory.FieldDataType, v)
	return u
}

// UpdateDataType sets the "data_type" field to the value that was provided on create.
func (u *ProviderSyncHistoryUpsert) UpdateDataType() *ProviderSyncHistoryUpsert {
	u.SetExcluded(providersynchistory.FieldDataType)
	return u
}

// SetStartedAt sets the "started_at" field.
func (u *ProviderSyncHistoryUpsert) SetStartedAt(v time.Time) *ProviderSyncHistoryUpsert {
	u.Set(providersynchistory.FieldStartedAt, v)
	return u
}

// UpdateStartedAt sets the "started_at" field to the value that was provided on create.
func (u *ProviderSyncHistoryUpsert) UpdateStartedAt() *ProviderSyncHistoryUpsert {
	u.SetExcluded(providersynchistory.FieldStartedAt)
	return u
}

// SetFinishedAt sets the "finished_at" field.
func (u *ProviderSyncHistoryUpsert) SetFinishedAt(v time.Time) *ProviderSyncHistoryUpsert {
	u.Set(providersynchistory.FieldFinishedAt, v)
	return u
}

// UpdateFinishedAt sets the "finished_at" field to the value that was provided on create.
func (u *ProviderSyncHistoryUpsert) UpdateFinishedAt() *ProviderSyncHistoryUpsert {
	u.SetExcluded(providersynchistory.FieldFinishedAt)
	return u
}

// SetNumMutations sets the "num_mutations" field.
func (u *ProviderSyncHistoryUpsert) SetNumMutations(v int) *ProviderSyncHistoryUpsert {
	u.Set(providersynchistory.FieldNumMutations, v)
	return u
}

// UpdateNumMutations sets the "num_mutations" field to the value that was provided on create.
func (u *ProviderSyncHistoryUpsert) UpdateNumMutations() *ProviderSyncHistoryUpsert {
	u.SetExcluded(providersynchistory.FieldNumMutations)
	return u
}

// AddNumMutations adds v to the "num_mutations" field.
func (u *ProviderSyncHistoryUpsert) AddNumMutations(v int) *ProviderSyncHistoryUpsert {
	u.Add(providersynchistory.FieldNumMutations, v)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.ProviderSyncHistory.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(providersynchistory.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *ProviderSyncHistoryUpsertOne) UpdateNewValues() *ProviderSyncHistoryUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(providersynchistory.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.ProviderSyncHistory.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *ProviderSyncHistoryUpsertOne) Ignore() *ProviderSyncHistoryUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *ProviderSyncHistoryUpsertOne) DoNothing() *ProviderSyncHistoryUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the ProviderSyncHistoryCreate.OnConflict
// documentation for more info.
func (u *ProviderSyncHistoryUpsertOne) Update(set func(*ProviderSyncHistoryUpsert)) *ProviderSyncHistoryUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&ProviderSyncHistoryUpsert{UpdateSet: update})
	}))
	return u
}

// SetDataType sets the "data_type" field.
func (u *ProviderSyncHistoryUpsertOne) SetDataType(v string) *ProviderSyncHistoryUpsertOne {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.SetDataType(v)
	})
}

// UpdateDataType sets the "data_type" field to the value that was provided on create.
func (u *ProviderSyncHistoryUpsertOne) UpdateDataType() *ProviderSyncHistoryUpsertOne {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.UpdateDataType()
	})
}

// SetStartedAt sets the "started_at" field.
func (u *ProviderSyncHistoryUpsertOne) SetStartedAt(v time.Time) *ProviderSyncHistoryUpsertOne {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.SetStartedAt(v)
	})
}

// UpdateStartedAt sets the "started_at" field to the value that was provided on create.
func (u *ProviderSyncHistoryUpsertOne) UpdateStartedAt() *ProviderSyncHistoryUpsertOne {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.UpdateStartedAt()
	})
}

// SetFinishedAt sets the "finished_at" field.
func (u *ProviderSyncHistoryUpsertOne) SetFinishedAt(v time.Time) *ProviderSyncHistoryUpsertOne {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.SetFinishedAt(v)
	})
}

// UpdateFinishedAt sets the "finished_at" field to the value that was provided on create.
func (u *ProviderSyncHistoryUpsertOne) UpdateFinishedAt() *ProviderSyncHistoryUpsertOne {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.UpdateFinishedAt()
	})
}

// SetNumMutations sets the "num_mutations" field.
func (u *ProviderSyncHistoryUpsertOne) SetNumMutations(v int) *ProviderSyncHistoryUpsertOne {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.SetNumMutations(v)
	})
}

// AddNumMutations adds v to the "num_mutations" field.
func (u *ProviderSyncHistoryUpsertOne) AddNumMutations(v int) *ProviderSyncHistoryUpsertOne {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.AddNumMutations(v)
	})
}

// UpdateNumMutations sets the "num_mutations" field to the value that was provided on create.
func (u *ProviderSyncHistoryUpsertOne) UpdateNumMutations() *ProviderSyncHistoryUpsertOne {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.UpdateNumMutations()
	})
}

// Exec executes the query.
func (u *ProviderSyncHistoryUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for ProviderSyncHistoryCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *ProviderSyncHistoryUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *ProviderSyncHistoryUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: ProviderSyncHistoryUpsertOne.ID is not supported by MySQL driver. Use ProviderSyncHistoryUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *ProviderSyncHistoryUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// ProviderSyncHistoryCreateBulk is the builder for creating many ProviderSyncHistory entities in bulk.
type ProviderSyncHistoryCreateBulk struct {
	config
	err      error
	builders []*ProviderSyncHistoryCreate
	conflict []sql.ConflictOption
}

// Save creates the ProviderSyncHistory entities in the database.
func (pshcb *ProviderSyncHistoryCreateBulk) Save(ctx context.Context) ([]*ProviderSyncHistory, error) {
	if pshcb.err != nil {
		return nil, pshcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(pshcb.builders))
	nodes := make([]*ProviderSyncHistory, len(pshcb.builders))
	mutators := make([]Mutator, len(pshcb.builders))
	for i := range pshcb.builders {
		func(i int, root context.Context) {
			builder := pshcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*ProviderSyncHistoryMutation)
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
					_, err = mutators[i+1].Mutate(root, pshcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = pshcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, pshcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, pshcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (pshcb *ProviderSyncHistoryCreateBulk) SaveX(ctx context.Context) []*ProviderSyncHistory {
	v, err := pshcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (pshcb *ProviderSyncHistoryCreateBulk) Exec(ctx context.Context) error {
	_, err := pshcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pshcb *ProviderSyncHistoryCreateBulk) ExecX(ctx context.Context) {
	if err := pshcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.ProviderSyncHistory.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.ProviderSyncHistoryUpsert) {
//			SetDataType(v+v).
//		}).
//		Exec(ctx)
func (pshcb *ProviderSyncHistoryCreateBulk) OnConflict(opts ...sql.ConflictOption) *ProviderSyncHistoryUpsertBulk {
	pshcb.conflict = opts
	return &ProviderSyncHistoryUpsertBulk{
		create: pshcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.ProviderSyncHistory.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (pshcb *ProviderSyncHistoryCreateBulk) OnConflictColumns(columns ...string) *ProviderSyncHistoryUpsertBulk {
	pshcb.conflict = append(pshcb.conflict, sql.ConflictColumns(columns...))
	return &ProviderSyncHistoryUpsertBulk{
		create: pshcb,
	}
}

// ProviderSyncHistoryUpsertBulk is the builder for "upsert"-ing
// a bulk of ProviderSyncHistory nodes.
type ProviderSyncHistoryUpsertBulk struct {
	create *ProviderSyncHistoryCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.ProviderSyncHistory.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(providersynchistory.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *ProviderSyncHistoryUpsertBulk) UpdateNewValues() *ProviderSyncHistoryUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(providersynchistory.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.ProviderSyncHistory.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *ProviderSyncHistoryUpsertBulk) Ignore() *ProviderSyncHistoryUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *ProviderSyncHistoryUpsertBulk) DoNothing() *ProviderSyncHistoryUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the ProviderSyncHistoryCreateBulk.OnConflict
// documentation for more info.
func (u *ProviderSyncHistoryUpsertBulk) Update(set func(*ProviderSyncHistoryUpsert)) *ProviderSyncHistoryUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&ProviderSyncHistoryUpsert{UpdateSet: update})
	}))
	return u
}

// SetDataType sets the "data_type" field.
func (u *ProviderSyncHistoryUpsertBulk) SetDataType(v string) *ProviderSyncHistoryUpsertBulk {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.SetDataType(v)
	})
}

// UpdateDataType sets the "data_type" field to the value that was provided on create.
func (u *ProviderSyncHistoryUpsertBulk) UpdateDataType() *ProviderSyncHistoryUpsertBulk {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.UpdateDataType()
	})
}

// SetStartedAt sets the "started_at" field.
func (u *ProviderSyncHistoryUpsertBulk) SetStartedAt(v time.Time) *ProviderSyncHistoryUpsertBulk {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.SetStartedAt(v)
	})
}

// UpdateStartedAt sets the "started_at" field to the value that was provided on create.
func (u *ProviderSyncHistoryUpsertBulk) UpdateStartedAt() *ProviderSyncHistoryUpsertBulk {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.UpdateStartedAt()
	})
}

// SetFinishedAt sets the "finished_at" field.
func (u *ProviderSyncHistoryUpsertBulk) SetFinishedAt(v time.Time) *ProviderSyncHistoryUpsertBulk {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.SetFinishedAt(v)
	})
}

// UpdateFinishedAt sets the "finished_at" field to the value that was provided on create.
func (u *ProviderSyncHistoryUpsertBulk) UpdateFinishedAt() *ProviderSyncHistoryUpsertBulk {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.UpdateFinishedAt()
	})
}

// SetNumMutations sets the "num_mutations" field.
func (u *ProviderSyncHistoryUpsertBulk) SetNumMutations(v int) *ProviderSyncHistoryUpsertBulk {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.SetNumMutations(v)
	})
}

// AddNumMutations adds v to the "num_mutations" field.
func (u *ProviderSyncHistoryUpsertBulk) AddNumMutations(v int) *ProviderSyncHistoryUpsertBulk {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.AddNumMutations(v)
	})
}

// UpdateNumMutations sets the "num_mutations" field to the value that was provided on create.
func (u *ProviderSyncHistoryUpsertBulk) UpdateNumMutations() *ProviderSyncHistoryUpsertBulk {
	return u.Update(func(s *ProviderSyncHistoryUpsert) {
		s.UpdateNumMutations()
	})
}

// Exec executes the query.
func (u *ProviderSyncHistoryUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the ProviderSyncHistoryCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for ProviderSyncHistoryCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *ProviderSyncHistoryUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}