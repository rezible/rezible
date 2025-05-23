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
	"github.com/rezible/rezible/ent/systemanalysisrelationship"
	"github.com/rezible/rezible/ent/systemcomponent"
	"github.com/rezible/rezible/ent/systemcomponentcontrol"
	"github.com/rezible/rezible/ent/systemrelationshipcontrolaction"
)

// SystemComponentControlCreate is the builder for creating a SystemComponentControl entity.
type SystemComponentControlCreate struct {
	config
	mutation *SystemComponentControlMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetComponentID sets the "component_id" field.
func (sccc *SystemComponentControlCreate) SetComponentID(u uuid.UUID) *SystemComponentControlCreate {
	sccc.mutation.SetComponentID(u)
	return sccc
}

// SetLabel sets the "label" field.
func (sccc *SystemComponentControlCreate) SetLabel(s string) *SystemComponentControlCreate {
	sccc.mutation.SetLabel(s)
	return sccc
}

// SetDescription sets the "description" field.
func (sccc *SystemComponentControlCreate) SetDescription(s string) *SystemComponentControlCreate {
	sccc.mutation.SetDescription(s)
	return sccc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (sccc *SystemComponentControlCreate) SetNillableDescription(s *string) *SystemComponentControlCreate {
	if s != nil {
		sccc.SetDescription(*s)
	}
	return sccc
}

// SetCreatedAt sets the "created_at" field.
func (sccc *SystemComponentControlCreate) SetCreatedAt(t time.Time) *SystemComponentControlCreate {
	sccc.mutation.SetCreatedAt(t)
	return sccc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (sccc *SystemComponentControlCreate) SetNillableCreatedAt(t *time.Time) *SystemComponentControlCreate {
	if t != nil {
		sccc.SetCreatedAt(*t)
	}
	return sccc
}

// SetID sets the "id" field.
func (sccc *SystemComponentControlCreate) SetID(u uuid.UUID) *SystemComponentControlCreate {
	sccc.mutation.SetID(u)
	return sccc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (sccc *SystemComponentControlCreate) SetNillableID(u *uuid.UUID) *SystemComponentControlCreate {
	if u != nil {
		sccc.SetID(*u)
	}
	return sccc
}

// SetComponent sets the "component" edge to the SystemComponent entity.
func (sccc *SystemComponentControlCreate) SetComponent(s *SystemComponent) *SystemComponentControlCreate {
	return sccc.SetComponentID(s.ID)
}

// AddRelationshipIDs adds the "relationships" edge to the SystemAnalysisRelationship entity by IDs.
func (sccc *SystemComponentControlCreate) AddRelationshipIDs(ids ...uuid.UUID) *SystemComponentControlCreate {
	sccc.mutation.AddRelationshipIDs(ids...)
	return sccc
}

// AddRelationships adds the "relationships" edges to the SystemAnalysisRelationship entity.
func (sccc *SystemComponentControlCreate) AddRelationships(s ...*SystemAnalysisRelationship) *SystemComponentControlCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sccc.AddRelationshipIDs(ids...)
}

// AddControlActionIDs adds the "control_actions" edge to the SystemRelationshipControlAction entity by IDs.
func (sccc *SystemComponentControlCreate) AddControlActionIDs(ids ...uuid.UUID) *SystemComponentControlCreate {
	sccc.mutation.AddControlActionIDs(ids...)
	return sccc
}

// AddControlActions adds the "control_actions" edges to the SystemRelationshipControlAction entity.
func (sccc *SystemComponentControlCreate) AddControlActions(s ...*SystemRelationshipControlAction) *SystemComponentControlCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sccc.AddControlActionIDs(ids...)
}

// Mutation returns the SystemComponentControlMutation object of the builder.
func (sccc *SystemComponentControlCreate) Mutation() *SystemComponentControlMutation {
	return sccc.mutation
}

// Save creates the SystemComponentControl in the database.
func (sccc *SystemComponentControlCreate) Save(ctx context.Context) (*SystemComponentControl, error) {
	sccc.defaults()
	return withHooks(ctx, sccc.sqlSave, sccc.mutation, sccc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (sccc *SystemComponentControlCreate) SaveX(ctx context.Context) *SystemComponentControl {
	v, err := sccc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (sccc *SystemComponentControlCreate) Exec(ctx context.Context) error {
	_, err := sccc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sccc *SystemComponentControlCreate) ExecX(ctx context.Context) {
	if err := sccc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (sccc *SystemComponentControlCreate) defaults() {
	if _, ok := sccc.mutation.CreatedAt(); !ok {
		v := systemcomponentcontrol.DefaultCreatedAt()
		sccc.mutation.SetCreatedAt(v)
	}
	if _, ok := sccc.mutation.ID(); !ok {
		v := systemcomponentcontrol.DefaultID()
		sccc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sccc *SystemComponentControlCreate) check() error {
	if _, ok := sccc.mutation.ComponentID(); !ok {
		return &ValidationError{Name: "component_id", err: errors.New(`ent: missing required field "SystemComponentControl.component_id"`)}
	}
	if _, ok := sccc.mutation.Label(); !ok {
		return &ValidationError{Name: "label", err: errors.New(`ent: missing required field "SystemComponentControl.label"`)}
	}
	if _, ok := sccc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "SystemComponentControl.created_at"`)}
	}
	if len(sccc.mutation.ComponentIDs()) == 0 {
		return &ValidationError{Name: "component", err: errors.New(`ent: missing required edge "SystemComponentControl.component"`)}
	}
	return nil
}

func (sccc *SystemComponentControlCreate) sqlSave(ctx context.Context) (*SystemComponentControl, error) {
	if err := sccc.check(); err != nil {
		return nil, err
	}
	_node, _spec := sccc.createSpec()
	if err := sqlgraph.CreateNode(ctx, sccc.driver, _spec); err != nil {
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
	sccc.mutation.id = &_node.ID
	sccc.mutation.done = true
	return _node, nil
}

func (sccc *SystemComponentControlCreate) createSpec() (*SystemComponentControl, *sqlgraph.CreateSpec) {
	var (
		_node = &SystemComponentControl{config: sccc.config}
		_spec = sqlgraph.NewCreateSpec(systemcomponentcontrol.Table, sqlgraph.NewFieldSpec(systemcomponentcontrol.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = sccc.conflict
	if id, ok := sccc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := sccc.mutation.Label(); ok {
		_spec.SetField(systemcomponentcontrol.FieldLabel, field.TypeString, value)
		_node.Label = value
	}
	if value, ok := sccc.mutation.Description(); ok {
		_spec.SetField(systemcomponentcontrol.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if value, ok := sccc.mutation.CreatedAt(); ok {
		_spec.SetField(systemcomponentcontrol.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if nodes := sccc.mutation.ComponentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentcontrol.ComponentTable,
			Columns: []string{systemcomponentcontrol.ComponentColumn},
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
	if nodes := sccc.mutation.RelationshipsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   systemcomponentcontrol.RelationshipsTable,
			Columns: systemcomponentcontrol.RelationshipsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysisrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		createE := &SystemRelationshipControlActionCreate{config: sccc.config, mutation: newSystemRelationshipControlActionMutation(sccc.config, OpCreate)}
		createE.defaults()
		_, specE := createE.createSpec()
		edge.Target.Fields = specE.Fields
		if specE.ID.Value != nil {
			edge.Target.Fields = append(edge.Target.Fields, specE.ID)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := sccc.mutation.ControlActionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemcomponentcontrol.ControlActionsTable,
			Columns: []string{systemcomponentcontrol.ControlActionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemrelationshipcontrolaction.FieldID, field.TypeUUID),
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
//	client.SystemComponentControl.Create().
//		SetComponentID(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.SystemComponentControlUpsert) {
//			SetComponentID(v+v).
//		}).
//		Exec(ctx)
func (sccc *SystemComponentControlCreate) OnConflict(opts ...sql.ConflictOption) *SystemComponentControlUpsertOne {
	sccc.conflict = opts
	return &SystemComponentControlUpsertOne{
		create: sccc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.SystemComponentControl.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (sccc *SystemComponentControlCreate) OnConflictColumns(columns ...string) *SystemComponentControlUpsertOne {
	sccc.conflict = append(sccc.conflict, sql.ConflictColumns(columns...))
	return &SystemComponentControlUpsertOne{
		create: sccc,
	}
}

type (
	// SystemComponentControlUpsertOne is the builder for "upsert"-ing
	//  one SystemComponentControl node.
	SystemComponentControlUpsertOne struct {
		create *SystemComponentControlCreate
	}

	// SystemComponentControlUpsert is the "OnConflict" setter.
	SystemComponentControlUpsert struct {
		*sql.UpdateSet
	}
)

// SetComponentID sets the "component_id" field.
func (u *SystemComponentControlUpsert) SetComponentID(v uuid.UUID) *SystemComponentControlUpsert {
	u.Set(systemcomponentcontrol.FieldComponentID, v)
	return u
}

// UpdateComponentID sets the "component_id" field to the value that was provided on create.
func (u *SystemComponentControlUpsert) UpdateComponentID() *SystemComponentControlUpsert {
	u.SetExcluded(systemcomponentcontrol.FieldComponentID)
	return u
}

// SetLabel sets the "label" field.
func (u *SystemComponentControlUpsert) SetLabel(v string) *SystemComponentControlUpsert {
	u.Set(systemcomponentcontrol.FieldLabel, v)
	return u
}

// UpdateLabel sets the "label" field to the value that was provided on create.
func (u *SystemComponentControlUpsert) UpdateLabel() *SystemComponentControlUpsert {
	u.SetExcluded(systemcomponentcontrol.FieldLabel)
	return u
}

// SetDescription sets the "description" field.
func (u *SystemComponentControlUpsert) SetDescription(v string) *SystemComponentControlUpsert {
	u.Set(systemcomponentcontrol.FieldDescription, v)
	return u
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *SystemComponentControlUpsert) UpdateDescription() *SystemComponentControlUpsert {
	u.SetExcluded(systemcomponentcontrol.FieldDescription)
	return u
}

// ClearDescription clears the value of the "description" field.
func (u *SystemComponentControlUpsert) ClearDescription() *SystemComponentControlUpsert {
	u.SetNull(systemcomponentcontrol.FieldDescription)
	return u
}

// SetCreatedAt sets the "created_at" field.
func (u *SystemComponentControlUpsert) SetCreatedAt(v time.Time) *SystemComponentControlUpsert {
	u.Set(systemcomponentcontrol.FieldCreatedAt, v)
	return u
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *SystemComponentControlUpsert) UpdateCreatedAt() *SystemComponentControlUpsert {
	u.SetExcluded(systemcomponentcontrol.FieldCreatedAt)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.SystemComponentControl.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(systemcomponentcontrol.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *SystemComponentControlUpsertOne) UpdateNewValues() *SystemComponentControlUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(systemcomponentcontrol.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.SystemComponentControl.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *SystemComponentControlUpsertOne) Ignore() *SystemComponentControlUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *SystemComponentControlUpsertOne) DoNothing() *SystemComponentControlUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the SystemComponentControlCreate.OnConflict
// documentation for more info.
func (u *SystemComponentControlUpsertOne) Update(set func(*SystemComponentControlUpsert)) *SystemComponentControlUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&SystemComponentControlUpsert{UpdateSet: update})
	}))
	return u
}

// SetComponentID sets the "component_id" field.
func (u *SystemComponentControlUpsertOne) SetComponentID(v uuid.UUID) *SystemComponentControlUpsertOne {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.SetComponentID(v)
	})
}

// UpdateComponentID sets the "component_id" field to the value that was provided on create.
func (u *SystemComponentControlUpsertOne) UpdateComponentID() *SystemComponentControlUpsertOne {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.UpdateComponentID()
	})
}

// SetLabel sets the "label" field.
func (u *SystemComponentControlUpsertOne) SetLabel(v string) *SystemComponentControlUpsertOne {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.SetLabel(v)
	})
}

// UpdateLabel sets the "label" field to the value that was provided on create.
func (u *SystemComponentControlUpsertOne) UpdateLabel() *SystemComponentControlUpsertOne {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.UpdateLabel()
	})
}

// SetDescription sets the "description" field.
func (u *SystemComponentControlUpsertOne) SetDescription(v string) *SystemComponentControlUpsertOne {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *SystemComponentControlUpsertOne) UpdateDescription() *SystemComponentControlUpsertOne {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *SystemComponentControlUpsertOne) ClearDescription() *SystemComponentControlUpsertOne {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.ClearDescription()
	})
}

// SetCreatedAt sets the "created_at" field.
func (u *SystemComponentControlUpsertOne) SetCreatedAt(v time.Time) *SystemComponentControlUpsertOne {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.SetCreatedAt(v)
	})
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *SystemComponentControlUpsertOne) UpdateCreatedAt() *SystemComponentControlUpsertOne {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.UpdateCreatedAt()
	})
}

// Exec executes the query.
func (u *SystemComponentControlUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for SystemComponentControlCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *SystemComponentControlUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *SystemComponentControlUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: SystemComponentControlUpsertOne.ID is not supported by MySQL driver. Use SystemComponentControlUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *SystemComponentControlUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// SystemComponentControlCreateBulk is the builder for creating many SystemComponentControl entities in bulk.
type SystemComponentControlCreateBulk struct {
	config
	err      error
	builders []*SystemComponentControlCreate
	conflict []sql.ConflictOption
}

// Save creates the SystemComponentControl entities in the database.
func (scccb *SystemComponentControlCreateBulk) Save(ctx context.Context) ([]*SystemComponentControl, error) {
	if scccb.err != nil {
		return nil, scccb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(scccb.builders))
	nodes := make([]*SystemComponentControl, len(scccb.builders))
	mutators := make([]Mutator, len(scccb.builders))
	for i := range scccb.builders {
		func(i int, root context.Context) {
			builder := scccb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SystemComponentControlMutation)
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
					_, err = mutators[i+1].Mutate(root, scccb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = scccb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, scccb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, scccb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (scccb *SystemComponentControlCreateBulk) SaveX(ctx context.Context) []*SystemComponentControl {
	v, err := scccb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (scccb *SystemComponentControlCreateBulk) Exec(ctx context.Context) error {
	_, err := scccb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (scccb *SystemComponentControlCreateBulk) ExecX(ctx context.Context) {
	if err := scccb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.SystemComponentControl.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.SystemComponentControlUpsert) {
//			SetComponentID(v+v).
//		}).
//		Exec(ctx)
func (scccb *SystemComponentControlCreateBulk) OnConflict(opts ...sql.ConflictOption) *SystemComponentControlUpsertBulk {
	scccb.conflict = opts
	return &SystemComponentControlUpsertBulk{
		create: scccb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.SystemComponentControl.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (scccb *SystemComponentControlCreateBulk) OnConflictColumns(columns ...string) *SystemComponentControlUpsertBulk {
	scccb.conflict = append(scccb.conflict, sql.ConflictColumns(columns...))
	return &SystemComponentControlUpsertBulk{
		create: scccb,
	}
}

// SystemComponentControlUpsertBulk is the builder for "upsert"-ing
// a bulk of SystemComponentControl nodes.
type SystemComponentControlUpsertBulk struct {
	create *SystemComponentControlCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.SystemComponentControl.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(systemcomponentcontrol.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *SystemComponentControlUpsertBulk) UpdateNewValues() *SystemComponentControlUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(systemcomponentcontrol.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.SystemComponentControl.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *SystemComponentControlUpsertBulk) Ignore() *SystemComponentControlUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *SystemComponentControlUpsertBulk) DoNothing() *SystemComponentControlUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the SystemComponentControlCreateBulk.OnConflict
// documentation for more info.
func (u *SystemComponentControlUpsertBulk) Update(set func(*SystemComponentControlUpsert)) *SystemComponentControlUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&SystemComponentControlUpsert{UpdateSet: update})
	}))
	return u
}

// SetComponentID sets the "component_id" field.
func (u *SystemComponentControlUpsertBulk) SetComponentID(v uuid.UUID) *SystemComponentControlUpsertBulk {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.SetComponentID(v)
	})
}

// UpdateComponentID sets the "component_id" field to the value that was provided on create.
func (u *SystemComponentControlUpsertBulk) UpdateComponentID() *SystemComponentControlUpsertBulk {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.UpdateComponentID()
	})
}

// SetLabel sets the "label" field.
func (u *SystemComponentControlUpsertBulk) SetLabel(v string) *SystemComponentControlUpsertBulk {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.SetLabel(v)
	})
}

// UpdateLabel sets the "label" field to the value that was provided on create.
func (u *SystemComponentControlUpsertBulk) UpdateLabel() *SystemComponentControlUpsertBulk {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.UpdateLabel()
	})
}

// SetDescription sets the "description" field.
func (u *SystemComponentControlUpsertBulk) SetDescription(v string) *SystemComponentControlUpsertBulk {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *SystemComponentControlUpsertBulk) UpdateDescription() *SystemComponentControlUpsertBulk {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *SystemComponentControlUpsertBulk) ClearDescription() *SystemComponentControlUpsertBulk {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.ClearDescription()
	})
}

// SetCreatedAt sets the "created_at" field.
func (u *SystemComponentControlUpsertBulk) SetCreatedAt(v time.Time) *SystemComponentControlUpsertBulk {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.SetCreatedAt(v)
	})
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *SystemComponentControlUpsertBulk) UpdateCreatedAt() *SystemComponentControlUpsertBulk {
	return u.Update(func(s *SystemComponentControlUpsert) {
		s.UpdateCreatedAt()
	})
}

// Exec executes the query.
func (u *SystemComponentControlUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the SystemComponentControlCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for SystemComponentControlCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *SystemComponentControlUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
