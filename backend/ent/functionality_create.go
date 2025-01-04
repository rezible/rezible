// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/functionality"
)

// FunctionalityCreate is the builder for creating a Functionality entity.
type FunctionalityCreate struct {
	config
	mutation *FunctionalityMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetName sets the "name" field.
func (fc *FunctionalityCreate) SetName(s string) *FunctionalityCreate {
	fc.mutation.SetName(s)
	return fc
}

// SetID sets the "id" field.
func (fc *FunctionalityCreate) SetID(u uuid.UUID) *FunctionalityCreate {
	fc.mutation.SetID(u)
	return fc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (fc *FunctionalityCreate) SetNillableID(u *uuid.UUID) *FunctionalityCreate {
	if u != nil {
		fc.SetID(*u)
	}
	return fc
}

// Mutation returns the FunctionalityMutation object of the builder.
func (fc *FunctionalityCreate) Mutation() *FunctionalityMutation {
	return fc.mutation
}

// Save creates the Functionality in the database.
func (fc *FunctionalityCreate) Save(ctx context.Context) (*Functionality, error) {
	fc.defaults()
	return withHooks(ctx, fc.sqlSave, fc.mutation, fc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (fc *FunctionalityCreate) SaveX(ctx context.Context) *Functionality {
	v, err := fc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (fc *FunctionalityCreate) Exec(ctx context.Context) error {
	_, err := fc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fc *FunctionalityCreate) ExecX(ctx context.Context) {
	if err := fc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (fc *FunctionalityCreate) defaults() {
	if _, ok := fc.mutation.ID(); !ok {
		v := functionality.DefaultID()
		fc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (fc *FunctionalityCreate) check() error {
	if _, ok := fc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`ent: missing required field "Functionality.name"`)}
	}
	return nil
}

func (fc *FunctionalityCreate) sqlSave(ctx context.Context) (*Functionality, error) {
	if err := fc.check(); err != nil {
		return nil, err
	}
	_node, _spec := fc.createSpec()
	if err := sqlgraph.CreateNode(ctx, fc.driver, _spec); err != nil {
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
	fc.mutation.id = &_node.ID
	fc.mutation.done = true
	return _node, nil
}

func (fc *FunctionalityCreate) createSpec() (*Functionality, *sqlgraph.CreateSpec) {
	var (
		_node = &Functionality{config: fc.config}
		_spec = sqlgraph.NewCreateSpec(functionality.Table, sqlgraph.NewFieldSpec(functionality.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = fc.conflict
	if id, ok := fc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := fc.mutation.Name(); ok {
		_spec.SetField(functionality.FieldName, field.TypeString, value)
		_node.Name = value
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Functionality.Create().
//		SetName(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.FunctionalityUpsert) {
//			SetName(v+v).
//		}).
//		Exec(ctx)
func (fc *FunctionalityCreate) OnConflict(opts ...sql.ConflictOption) *FunctionalityUpsertOne {
	fc.conflict = opts
	return &FunctionalityUpsertOne{
		create: fc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Functionality.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (fc *FunctionalityCreate) OnConflictColumns(columns ...string) *FunctionalityUpsertOne {
	fc.conflict = append(fc.conflict, sql.ConflictColumns(columns...))
	return &FunctionalityUpsertOne{
		create: fc,
	}
}

type (
	// FunctionalityUpsertOne is the builder for "upsert"-ing
	//  one Functionality node.
	FunctionalityUpsertOne struct {
		create *FunctionalityCreate
	}

	// FunctionalityUpsert is the "OnConflict" setter.
	FunctionalityUpsert struct {
		*sql.UpdateSet
	}
)

// SetName sets the "name" field.
func (u *FunctionalityUpsert) SetName(v string) *FunctionalityUpsert {
	u.Set(functionality.FieldName, v)
	return u
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *FunctionalityUpsert) UpdateName() *FunctionalityUpsert {
	u.SetExcluded(functionality.FieldName)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.Functionality.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(functionality.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *FunctionalityUpsertOne) UpdateNewValues() *FunctionalityUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(functionality.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Functionality.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *FunctionalityUpsertOne) Ignore() *FunctionalityUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *FunctionalityUpsertOne) DoNothing() *FunctionalityUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the FunctionalityCreate.OnConflict
// documentation for more info.
func (u *FunctionalityUpsertOne) Update(set func(*FunctionalityUpsert)) *FunctionalityUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&FunctionalityUpsert{UpdateSet: update})
	}))
	return u
}

// SetName sets the "name" field.
func (u *FunctionalityUpsertOne) SetName(v string) *FunctionalityUpsertOne {
	return u.Update(func(s *FunctionalityUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *FunctionalityUpsertOne) UpdateName() *FunctionalityUpsertOne {
	return u.Update(func(s *FunctionalityUpsert) {
		s.UpdateName()
	})
}

// Exec executes the query.
func (u *FunctionalityUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for FunctionalityCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *FunctionalityUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *FunctionalityUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: FunctionalityUpsertOne.ID is not supported by MySQL driver. Use FunctionalityUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *FunctionalityUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// FunctionalityCreateBulk is the builder for creating many Functionality entities in bulk.
type FunctionalityCreateBulk struct {
	config
	err      error
	builders []*FunctionalityCreate
	conflict []sql.ConflictOption
}

// Save creates the Functionality entities in the database.
func (fcb *FunctionalityCreateBulk) Save(ctx context.Context) ([]*Functionality, error) {
	if fcb.err != nil {
		return nil, fcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(fcb.builders))
	nodes := make([]*Functionality, len(fcb.builders))
	mutators := make([]Mutator, len(fcb.builders))
	for i := range fcb.builders {
		func(i int, root context.Context) {
			builder := fcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*FunctionalityMutation)
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
					_, err = mutators[i+1].Mutate(root, fcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = fcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, fcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, fcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (fcb *FunctionalityCreateBulk) SaveX(ctx context.Context) []*Functionality {
	v, err := fcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (fcb *FunctionalityCreateBulk) Exec(ctx context.Context) error {
	_, err := fcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fcb *FunctionalityCreateBulk) ExecX(ctx context.Context) {
	if err := fcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Functionality.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.FunctionalityUpsert) {
//			SetName(v+v).
//		}).
//		Exec(ctx)
func (fcb *FunctionalityCreateBulk) OnConflict(opts ...sql.ConflictOption) *FunctionalityUpsertBulk {
	fcb.conflict = opts
	return &FunctionalityUpsertBulk{
		create: fcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Functionality.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (fcb *FunctionalityCreateBulk) OnConflictColumns(columns ...string) *FunctionalityUpsertBulk {
	fcb.conflict = append(fcb.conflict, sql.ConflictColumns(columns...))
	return &FunctionalityUpsertBulk{
		create: fcb,
	}
}

// FunctionalityUpsertBulk is the builder for "upsert"-ing
// a bulk of Functionality nodes.
type FunctionalityUpsertBulk struct {
	create *FunctionalityCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.Functionality.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(functionality.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *FunctionalityUpsertBulk) UpdateNewValues() *FunctionalityUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(functionality.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Functionality.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *FunctionalityUpsertBulk) Ignore() *FunctionalityUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *FunctionalityUpsertBulk) DoNothing() *FunctionalityUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the FunctionalityCreateBulk.OnConflict
// documentation for more info.
func (u *FunctionalityUpsertBulk) Update(set func(*FunctionalityUpsert)) *FunctionalityUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&FunctionalityUpsert{UpdateSet: update})
	}))
	return u
}

// SetName sets the "name" field.
func (u *FunctionalityUpsertBulk) SetName(v string) *FunctionalityUpsertBulk {
	return u.Update(func(s *FunctionalityUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *FunctionalityUpsertBulk) UpdateName() *FunctionalityUpsertBulk {
	return u.Update(func(s *FunctionalityUpsert) {
		s.UpdateName()
	})
}

// Exec executes the query.
func (u *FunctionalityUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the FunctionalityCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for FunctionalityCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *FunctionalityUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
