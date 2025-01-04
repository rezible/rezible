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
	"github.com/rezible/rezible/ent/systemcomponentfeedbackrelationship"
)

// SystemComponentFeedbackRelationshipCreate is the builder for creating a SystemComponentFeedbackRelationship entity.
type SystemComponentFeedbackRelationshipCreate struct {
	config
	mutation *SystemComponentFeedbackRelationshipMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetSourceID sets the "source_id" field.
func (scfrc *SystemComponentFeedbackRelationshipCreate) SetSourceID(u uuid.UUID) *SystemComponentFeedbackRelationshipCreate {
	scfrc.mutation.SetSourceID(u)
	return scfrc
}

// SetTargetID sets the "target_id" field.
func (scfrc *SystemComponentFeedbackRelationshipCreate) SetTargetID(u uuid.UUID) *SystemComponentFeedbackRelationshipCreate {
	scfrc.mutation.SetTargetID(u)
	return scfrc
}

// SetType sets the "type" field.
func (scfrc *SystemComponentFeedbackRelationshipCreate) SetType(s string) *SystemComponentFeedbackRelationshipCreate {
	scfrc.mutation.SetType(s)
	return scfrc
}

// SetDescription sets the "description" field.
func (scfrc *SystemComponentFeedbackRelationshipCreate) SetDescription(s string) *SystemComponentFeedbackRelationshipCreate {
	scfrc.mutation.SetDescription(s)
	return scfrc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (scfrc *SystemComponentFeedbackRelationshipCreate) SetNillableDescription(s *string) *SystemComponentFeedbackRelationshipCreate {
	if s != nil {
		scfrc.SetDescription(*s)
	}
	return scfrc
}

// SetCreatedAt sets the "created_at" field.
func (scfrc *SystemComponentFeedbackRelationshipCreate) SetCreatedAt(t time.Time) *SystemComponentFeedbackRelationshipCreate {
	scfrc.mutation.SetCreatedAt(t)
	return scfrc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (scfrc *SystemComponentFeedbackRelationshipCreate) SetNillableCreatedAt(t *time.Time) *SystemComponentFeedbackRelationshipCreate {
	if t != nil {
		scfrc.SetCreatedAt(*t)
	}
	return scfrc
}

// SetID sets the "id" field.
func (scfrc *SystemComponentFeedbackRelationshipCreate) SetID(u uuid.UUID) *SystemComponentFeedbackRelationshipCreate {
	scfrc.mutation.SetID(u)
	return scfrc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (scfrc *SystemComponentFeedbackRelationshipCreate) SetNillableID(u *uuid.UUID) *SystemComponentFeedbackRelationshipCreate {
	if u != nil {
		scfrc.SetID(*u)
	}
	return scfrc
}

// SetSource sets the "source" edge to the SystemComponent entity.
func (scfrc *SystemComponentFeedbackRelationshipCreate) SetSource(s *SystemComponent) *SystemComponentFeedbackRelationshipCreate {
	return scfrc.SetSourceID(s.ID)
}

// SetTarget sets the "target" edge to the SystemComponent entity.
func (scfrc *SystemComponentFeedbackRelationshipCreate) SetTarget(s *SystemComponent) *SystemComponentFeedbackRelationshipCreate {
	return scfrc.SetTargetID(s.ID)
}

// Mutation returns the SystemComponentFeedbackRelationshipMutation object of the builder.
func (scfrc *SystemComponentFeedbackRelationshipCreate) Mutation() *SystemComponentFeedbackRelationshipMutation {
	return scfrc.mutation
}

// Save creates the SystemComponentFeedbackRelationship in the database.
func (scfrc *SystemComponentFeedbackRelationshipCreate) Save(ctx context.Context) (*SystemComponentFeedbackRelationship, error) {
	scfrc.defaults()
	return withHooks(ctx, scfrc.sqlSave, scfrc.mutation, scfrc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (scfrc *SystemComponentFeedbackRelationshipCreate) SaveX(ctx context.Context) *SystemComponentFeedbackRelationship {
	v, err := scfrc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (scfrc *SystemComponentFeedbackRelationshipCreate) Exec(ctx context.Context) error {
	_, err := scfrc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (scfrc *SystemComponentFeedbackRelationshipCreate) ExecX(ctx context.Context) {
	if err := scfrc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (scfrc *SystemComponentFeedbackRelationshipCreate) defaults() {
	if _, ok := scfrc.mutation.CreatedAt(); !ok {
		v := systemcomponentfeedbackrelationship.DefaultCreatedAt()
		scfrc.mutation.SetCreatedAt(v)
	}
	if _, ok := scfrc.mutation.ID(); !ok {
		v := systemcomponentfeedbackrelationship.DefaultID()
		scfrc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (scfrc *SystemComponentFeedbackRelationshipCreate) check() error {
	if _, ok := scfrc.mutation.SourceID(); !ok {
		return &ValidationError{Name: "source_id", err: errors.New(`ent: missing required field "SystemComponentFeedbackRelationship.source_id"`)}
	}
	if _, ok := scfrc.mutation.TargetID(); !ok {
		return &ValidationError{Name: "target_id", err: errors.New(`ent: missing required field "SystemComponentFeedbackRelationship.target_id"`)}
	}
	if _, ok := scfrc.mutation.GetType(); !ok {
		return &ValidationError{Name: "type", err: errors.New(`ent: missing required field "SystemComponentFeedbackRelationship.type"`)}
	}
	if v, ok := scfrc.mutation.GetType(); ok {
		if err := systemcomponentfeedbackrelationship.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "SystemComponentFeedbackRelationship.type": %w`, err)}
		}
	}
	if _, ok := scfrc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "SystemComponentFeedbackRelationship.created_at"`)}
	}
	if len(scfrc.mutation.SourceIDs()) == 0 {
		return &ValidationError{Name: "source", err: errors.New(`ent: missing required edge "SystemComponentFeedbackRelationship.source"`)}
	}
	if len(scfrc.mutation.TargetIDs()) == 0 {
		return &ValidationError{Name: "target", err: errors.New(`ent: missing required edge "SystemComponentFeedbackRelationship.target"`)}
	}
	return nil
}

func (scfrc *SystemComponentFeedbackRelationshipCreate) sqlSave(ctx context.Context) (*SystemComponentFeedbackRelationship, error) {
	if err := scfrc.check(); err != nil {
		return nil, err
	}
	_node, _spec := scfrc.createSpec()
	if err := sqlgraph.CreateNode(ctx, scfrc.driver, _spec); err != nil {
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
	scfrc.mutation.id = &_node.ID
	scfrc.mutation.done = true
	return _node, nil
}

func (scfrc *SystemComponentFeedbackRelationshipCreate) createSpec() (*SystemComponentFeedbackRelationship, *sqlgraph.CreateSpec) {
	var (
		_node = &SystemComponentFeedbackRelationship{config: scfrc.config}
		_spec = sqlgraph.NewCreateSpec(systemcomponentfeedbackrelationship.Table, sqlgraph.NewFieldSpec(systemcomponentfeedbackrelationship.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = scfrc.conflict
	if id, ok := scfrc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := scfrc.mutation.GetType(); ok {
		_spec.SetField(systemcomponentfeedbackrelationship.FieldType, field.TypeString, value)
		_node.Type = value
	}
	if value, ok := scfrc.mutation.Description(); ok {
		_spec.SetField(systemcomponentfeedbackrelationship.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if value, ok := scfrc.mutation.CreatedAt(); ok {
		_spec.SetField(systemcomponentfeedbackrelationship.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if nodes := scfrc.mutation.SourceIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentfeedbackrelationship.SourceTable,
			Columns: []string{systemcomponentfeedbackrelationship.SourceColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.SourceID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := scfrc.mutation.TargetIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentfeedbackrelationship.TargetTable,
			Columns: []string{systemcomponentfeedbackrelationship.TargetColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.TargetID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.SystemComponentFeedbackRelationship.Create().
//		SetSourceID(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.SystemComponentFeedbackRelationshipUpsert) {
//			SetSourceID(v+v).
//		}).
//		Exec(ctx)
func (scfrc *SystemComponentFeedbackRelationshipCreate) OnConflict(opts ...sql.ConflictOption) *SystemComponentFeedbackRelationshipUpsertOne {
	scfrc.conflict = opts
	return &SystemComponentFeedbackRelationshipUpsertOne{
		create: scfrc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.SystemComponentFeedbackRelationship.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (scfrc *SystemComponentFeedbackRelationshipCreate) OnConflictColumns(columns ...string) *SystemComponentFeedbackRelationshipUpsertOne {
	scfrc.conflict = append(scfrc.conflict, sql.ConflictColumns(columns...))
	return &SystemComponentFeedbackRelationshipUpsertOne{
		create: scfrc,
	}
}

type (
	// SystemComponentFeedbackRelationshipUpsertOne is the builder for "upsert"-ing
	//  one SystemComponentFeedbackRelationship node.
	SystemComponentFeedbackRelationshipUpsertOne struct {
		create *SystemComponentFeedbackRelationshipCreate
	}

	// SystemComponentFeedbackRelationshipUpsert is the "OnConflict" setter.
	SystemComponentFeedbackRelationshipUpsert struct {
		*sql.UpdateSet
	}
)

// SetSourceID sets the "source_id" field.
func (u *SystemComponentFeedbackRelationshipUpsert) SetSourceID(v uuid.UUID) *SystemComponentFeedbackRelationshipUpsert {
	u.Set(systemcomponentfeedbackrelationship.FieldSourceID, v)
	return u
}

// UpdateSourceID sets the "source_id" field to the value that was provided on create.
func (u *SystemComponentFeedbackRelationshipUpsert) UpdateSourceID() *SystemComponentFeedbackRelationshipUpsert {
	u.SetExcluded(systemcomponentfeedbackrelationship.FieldSourceID)
	return u
}

// SetTargetID sets the "target_id" field.
func (u *SystemComponentFeedbackRelationshipUpsert) SetTargetID(v uuid.UUID) *SystemComponentFeedbackRelationshipUpsert {
	u.Set(systemcomponentfeedbackrelationship.FieldTargetID, v)
	return u
}

// UpdateTargetID sets the "target_id" field to the value that was provided on create.
func (u *SystemComponentFeedbackRelationshipUpsert) UpdateTargetID() *SystemComponentFeedbackRelationshipUpsert {
	u.SetExcluded(systemcomponentfeedbackrelationship.FieldTargetID)
	return u
}

// SetType sets the "type" field.
func (u *SystemComponentFeedbackRelationshipUpsert) SetType(v string) *SystemComponentFeedbackRelationshipUpsert {
	u.Set(systemcomponentfeedbackrelationship.FieldType, v)
	return u
}

// UpdateType sets the "type" field to the value that was provided on create.
func (u *SystemComponentFeedbackRelationshipUpsert) UpdateType() *SystemComponentFeedbackRelationshipUpsert {
	u.SetExcluded(systemcomponentfeedbackrelationship.FieldType)
	return u
}

// SetDescription sets the "description" field.
func (u *SystemComponentFeedbackRelationshipUpsert) SetDescription(v string) *SystemComponentFeedbackRelationshipUpsert {
	u.Set(systemcomponentfeedbackrelationship.FieldDescription, v)
	return u
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *SystemComponentFeedbackRelationshipUpsert) UpdateDescription() *SystemComponentFeedbackRelationshipUpsert {
	u.SetExcluded(systemcomponentfeedbackrelationship.FieldDescription)
	return u
}

// ClearDescription clears the value of the "description" field.
func (u *SystemComponentFeedbackRelationshipUpsert) ClearDescription() *SystemComponentFeedbackRelationshipUpsert {
	u.SetNull(systemcomponentfeedbackrelationship.FieldDescription)
	return u
}

// SetCreatedAt sets the "created_at" field.
func (u *SystemComponentFeedbackRelationshipUpsert) SetCreatedAt(v time.Time) *SystemComponentFeedbackRelationshipUpsert {
	u.Set(systemcomponentfeedbackrelationship.FieldCreatedAt, v)
	return u
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *SystemComponentFeedbackRelationshipUpsert) UpdateCreatedAt() *SystemComponentFeedbackRelationshipUpsert {
	u.SetExcluded(systemcomponentfeedbackrelationship.FieldCreatedAt)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.SystemComponentFeedbackRelationship.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(systemcomponentfeedbackrelationship.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *SystemComponentFeedbackRelationshipUpsertOne) UpdateNewValues() *SystemComponentFeedbackRelationshipUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(systemcomponentfeedbackrelationship.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.SystemComponentFeedbackRelationship.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *SystemComponentFeedbackRelationshipUpsertOne) Ignore() *SystemComponentFeedbackRelationshipUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *SystemComponentFeedbackRelationshipUpsertOne) DoNothing() *SystemComponentFeedbackRelationshipUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the SystemComponentFeedbackRelationshipCreate.OnConflict
// documentation for more info.
func (u *SystemComponentFeedbackRelationshipUpsertOne) Update(set func(*SystemComponentFeedbackRelationshipUpsert)) *SystemComponentFeedbackRelationshipUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&SystemComponentFeedbackRelationshipUpsert{UpdateSet: update})
	}))
	return u
}

// SetSourceID sets the "source_id" field.
func (u *SystemComponentFeedbackRelationshipUpsertOne) SetSourceID(v uuid.UUID) *SystemComponentFeedbackRelationshipUpsertOne {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.SetSourceID(v)
	})
}

// UpdateSourceID sets the "source_id" field to the value that was provided on create.
func (u *SystemComponentFeedbackRelationshipUpsertOne) UpdateSourceID() *SystemComponentFeedbackRelationshipUpsertOne {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.UpdateSourceID()
	})
}

// SetTargetID sets the "target_id" field.
func (u *SystemComponentFeedbackRelationshipUpsertOne) SetTargetID(v uuid.UUID) *SystemComponentFeedbackRelationshipUpsertOne {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.SetTargetID(v)
	})
}

// UpdateTargetID sets the "target_id" field to the value that was provided on create.
func (u *SystemComponentFeedbackRelationshipUpsertOne) UpdateTargetID() *SystemComponentFeedbackRelationshipUpsertOne {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.UpdateTargetID()
	})
}

// SetType sets the "type" field.
func (u *SystemComponentFeedbackRelationshipUpsertOne) SetType(v string) *SystemComponentFeedbackRelationshipUpsertOne {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.SetType(v)
	})
}

// UpdateType sets the "type" field to the value that was provided on create.
func (u *SystemComponentFeedbackRelationshipUpsertOne) UpdateType() *SystemComponentFeedbackRelationshipUpsertOne {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.UpdateType()
	})
}

// SetDescription sets the "description" field.
func (u *SystemComponentFeedbackRelationshipUpsertOne) SetDescription(v string) *SystemComponentFeedbackRelationshipUpsertOne {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *SystemComponentFeedbackRelationshipUpsertOne) UpdateDescription() *SystemComponentFeedbackRelationshipUpsertOne {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *SystemComponentFeedbackRelationshipUpsertOne) ClearDescription() *SystemComponentFeedbackRelationshipUpsertOne {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.ClearDescription()
	})
}

// SetCreatedAt sets the "created_at" field.
func (u *SystemComponentFeedbackRelationshipUpsertOne) SetCreatedAt(v time.Time) *SystemComponentFeedbackRelationshipUpsertOne {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.SetCreatedAt(v)
	})
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *SystemComponentFeedbackRelationshipUpsertOne) UpdateCreatedAt() *SystemComponentFeedbackRelationshipUpsertOne {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.UpdateCreatedAt()
	})
}

// Exec executes the query.
func (u *SystemComponentFeedbackRelationshipUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for SystemComponentFeedbackRelationshipCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *SystemComponentFeedbackRelationshipUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *SystemComponentFeedbackRelationshipUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: SystemComponentFeedbackRelationshipUpsertOne.ID is not supported by MySQL driver. Use SystemComponentFeedbackRelationshipUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *SystemComponentFeedbackRelationshipUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// SystemComponentFeedbackRelationshipCreateBulk is the builder for creating many SystemComponentFeedbackRelationship entities in bulk.
type SystemComponentFeedbackRelationshipCreateBulk struct {
	config
	err      error
	builders []*SystemComponentFeedbackRelationshipCreate
	conflict []sql.ConflictOption
}

// Save creates the SystemComponentFeedbackRelationship entities in the database.
func (scfrcb *SystemComponentFeedbackRelationshipCreateBulk) Save(ctx context.Context) ([]*SystemComponentFeedbackRelationship, error) {
	if scfrcb.err != nil {
		return nil, scfrcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(scfrcb.builders))
	nodes := make([]*SystemComponentFeedbackRelationship, len(scfrcb.builders))
	mutators := make([]Mutator, len(scfrcb.builders))
	for i := range scfrcb.builders {
		func(i int, root context.Context) {
			builder := scfrcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SystemComponentFeedbackRelationshipMutation)
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
					_, err = mutators[i+1].Mutate(root, scfrcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = scfrcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, scfrcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, scfrcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (scfrcb *SystemComponentFeedbackRelationshipCreateBulk) SaveX(ctx context.Context) []*SystemComponentFeedbackRelationship {
	v, err := scfrcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (scfrcb *SystemComponentFeedbackRelationshipCreateBulk) Exec(ctx context.Context) error {
	_, err := scfrcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (scfrcb *SystemComponentFeedbackRelationshipCreateBulk) ExecX(ctx context.Context) {
	if err := scfrcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.SystemComponentFeedbackRelationship.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.SystemComponentFeedbackRelationshipUpsert) {
//			SetSourceID(v+v).
//		}).
//		Exec(ctx)
func (scfrcb *SystemComponentFeedbackRelationshipCreateBulk) OnConflict(opts ...sql.ConflictOption) *SystemComponentFeedbackRelationshipUpsertBulk {
	scfrcb.conflict = opts
	return &SystemComponentFeedbackRelationshipUpsertBulk{
		create: scfrcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.SystemComponentFeedbackRelationship.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (scfrcb *SystemComponentFeedbackRelationshipCreateBulk) OnConflictColumns(columns ...string) *SystemComponentFeedbackRelationshipUpsertBulk {
	scfrcb.conflict = append(scfrcb.conflict, sql.ConflictColumns(columns...))
	return &SystemComponentFeedbackRelationshipUpsertBulk{
		create: scfrcb,
	}
}

// SystemComponentFeedbackRelationshipUpsertBulk is the builder for "upsert"-ing
// a bulk of SystemComponentFeedbackRelationship nodes.
type SystemComponentFeedbackRelationshipUpsertBulk struct {
	create *SystemComponentFeedbackRelationshipCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.SystemComponentFeedbackRelationship.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(systemcomponentfeedbackrelationship.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *SystemComponentFeedbackRelationshipUpsertBulk) UpdateNewValues() *SystemComponentFeedbackRelationshipUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(systemcomponentfeedbackrelationship.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.SystemComponentFeedbackRelationship.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *SystemComponentFeedbackRelationshipUpsertBulk) Ignore() *SystemComponentFeedbackRelationshipUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *SystemComponentFeedbackRelationshipUpsertBulk) DoNothing() *SystemComponentFeedbackRelationshipUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the SystemComponentFeedbackRelationshipCreateBulk.OnConflict
// documentation for more info.
func (u *SystemComponentFeedbackRelationshipUpsertBulk) Update(set func(*SystemComponentFeedbackRelationshipUpsert)) *SystemComponentFeedbackRelationshipUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&SystemComponentFeedbackRelationshipUpsert{UpdateSet: update})
	}))
	return u
}

// SetSourceID sets the "source_id" field.
func (u *SystemComponentFeedbackRelationshipUpsertBulk) SetSourceID(v uuid.UUID) *SystemComponentFeedbackRelationshipUpsertBulk {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.SetSourceID(v)
	})
}

// UpdateSourceID sets the "source_id" field to the value that was provided on create.
func (u *SystemComponentFeedbackRelationshipUpsertBulk) UpdateSourceID() *SystemComponentFeedbackRelationshipUpsertBulk {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.UpdateSourceID()
	})
}

// SetTargetID sets the "target_id" field.
func (u *SystemComponentFeedbackRelationshipUpsertBulk) SetTargetID(v uuid.UUID) *SystemComponentFeedbackRelationshipUpsertBulk {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.SetTargetID(v)
	})
}

// UpdateTargetID sets the "target_id" field to the value that was provided on create.
func (u *SystemComponentFeedbackRelationshipUpsertBulk) UpdateTargetID() *SystemComponentFeedbackRelationshipUpsertBulk {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.UpdateTargetID()
	})
}

// SetType sets the "type" field.
func (u *SystemComponentFeedbackRelationshipUpsertBulk) SetType(v string) *SystemComponentFeedbackRelationshipUpsertBulk {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.SetType(v)
	})
}

// UpdateType sets the "type" field to the value that was provided on create.
func (u *SystemComponentFeedbackRelationshipUpsertBulk) UpdateType() *SystemComponentFeedbackRelationshipUpsertBulk {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.UpdateType()
	})
}

// SetDescription sets the "description" field.
func (u *SystemComponentFeedbackRelationshipUpsertBulk) SetDescription(v string) *SystemComponentFeedbackRelationshipUpsertBulk {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *SystemComponentFeedbackRelationshipUpsertBulk) UpdateDescription() *SystemComponentFeedbackRelationshipUpsertBulk {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *SystemComponentFeedbackRelationshipUpsertBulk) ClearDescription() *SystemComponentFeedbackRelationshipUpsertBulk {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.ClearDescription()
	})
}

// SetCreatedAt sets the "created_at" field.
func (u *SystemComponentFeedbackRelationshipUpsertBulk) SetCreatedAt(v time.Time) *SystemComponentFeedbackRelationshipUpsertBulk {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.SetCreatedAt(v)
	})
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *SystemComponentFeedbackRelationshipUpsertBulk) UpdateCreatedAt() *SystemComponentFeedbackRelationshipUpsertBulk {
	return u.Update(func(s *SystemComponentFeedbackRelationshipUpsert) {
		s.UpdateCreatedAt()
	})
}

// Exec executes the query.
func (u *SystemComponentFeedbackRelationshipUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the SystemComponentFeedbackRelationshipCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for SystemComponentFeedbackRelationshipCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *SystemComponentFeedbackRelationshipUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
