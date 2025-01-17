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
	"github.com/rezible/rezible/ent/systemcomponentcontrol"
	"github.com/rezible/rezible/ent/systemcomponentsignal"
	"github.com/rezible/rezible/ent/systemrelationship"
	"github.com/rezible/rezible/ent/systemrelationshipcontrolaction"
	"github.com/rezible/rezible/ent/systemrelationshipfeedback"
)

// SystemRelationshipCreate is the builder for creating a SystemRelationship entity.
type SystemRelationshipCreate struct {
	config
	mutation *SystemRelationshipMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetSourceComponentID sets the "source_component_id" field.
func (src *SystemRelationshipCreate) SetSourceComponentID(u uuid.UUID) *SystemRelationshipCreate {
	src.mutation.SetSourceComponentID(u)
	return src
}

// SetTargetComponentID sets the "target_component_id" field.
func (src *SystemRelationshipCreate) SetTargetComponentID(u uuid.UUID) *SystemRelationshipCreate {
	src.mutation.SetTargetComponentID(u)
	return src
}

// SetDescription sets the "description" field.
func (src *SystemRelationshipCreate) SetDescription(s string) *SystemRelationshipCreate {
	src.mutation.SetDescription(s)
	return src
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (src *SystemRelationshipCreate) SetNillableDescription(s *string) *SystemRelationshipCreate {
	if s != nil {
		src.SetDescription(*s)
	}
	return src
}

// SetCreatedAt sets the "created_at" field.
func (src *SystemRelationshipCreate) SetCreatedAt(t time.Time) *SystemRelationshipCreate {
	src.mutation.SetCreatedAt(t)
	return src
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (src *SystemRelationshipCreate) SetNillableCreatedAt(t *time.Time) *SystemRelationshipCreate {
	if t != nil {
		src.SetCreatedAt(*t)
	}
	return src
}

// SetID sets the "id" field.
func (src *SystemRelationshipCreate) SetID(u uuid.UUID) *SystemRelationshipCreate {
	src.mutation.SetID(u)
	return src
}

// SetNillableID sets the "id" field if the given value is not nil.
func (src *SystemRelationshipCreate) SetNillableID(u *uuid.UUID) *SystemRelationshipCreate {
	if u != nil {
		src.SetID(*u)
	}
	return src
}

// SetSourceComponent sets the "source_component" edge to the SystemComponent entity.
func (src *SystemRelationshipCreate) SetSourceComponent(s *SystemComponent) *SystemRelationshipCreate {
	return src.SetSourceComponentID(s.ID)
}

// SetTargetComponent sets the "target_component" edge to the SystemComponent entity.
func (src *SystemRelationshipCreate) SetTargetComponent(s *SystemComponent) *SystemRelationshipCreate {
	return src.SetTargetComponentID(s.ID)
}

// AddControlIDs adds the "controls" edge to the SystemComponentControl entity by IDs.
func (src *SystemRelationshipCreate) AddControlIDs(ids ...uuid.UUID) *SystemRelationshipCreate {
	src.mutation.AddControlIDs(ids...)
	return src
}

// AddControls adds the "controls" edges to the SystemComponentControl entity.
func (src *SystemRelationshipCreate) AddControls(s ...*SystemComponentControl) *SystemRelationshipCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return src.AddControlIDs(ids...)
}

// AddSignalIDs adds the "signals" edge to the SystemComponentSignal entity by IDs.
func (src *SystemRelationshipCreate) AddSignalIDs(ids ...uuid.UUID) *SystemRelationshipCreate {
	src.mutation.AddSignalIDs(ids...)
	return src
}

// AddSignals adds the "signals" edges to the SystemComponentSignal entity.
func (src *SystemRelationshipCreate) AddSignals(s ...*SystemComponentSignal) *SystemRelationshipCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return src.AddSignalIDs(ids...)
}

// AddControlActionIDs adds the "control_actions" edge to the SystemRelationshipControlAction entity by IDs.
func (src *SystemRelationshipCreate) AddControlActionIDs(ids ...uuid.UUID) *SystemRelationshipCreate {
	src.mutation.AddControlActionIDs(ids...)
	return src
}

// AddControlActions adds the "control_actions" edges to the SystemRelationshipControlAction entity.
func (src *SystemRelationshipCreate) AddControlActions(s ...*SystemRelationshipControlAction) *SystemRelationshipCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return src.AddControlActionIDs(ids...)
}

// AddFeedbackIDs adds the "feedback" edge to the SystemRelationshipFeedback entity by IDs.
func (src *SystemRelationshipCreate) AddFeedbackIDs(ids ...uuid.UUID) *SystemRelationshipCreate {
	src.mutation.AddFeedbackIDs(ids...)
	return src
}

// AddFeedback adds the "feedback" edges to the SystemRelationshipFeedback entity.
func (src *SystemRelationshipCreate) AddFeedback(s ...*SystemRelationshipFeedback) *SystemRelationshipCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return src.AddFeedbackIDs(ids...)
}

// Mutation returns the SystemRelationshipMutation object of the builder.
func (src *SystemRelationshipCreate) Mutation() *SystemRelationshipMutation {
	return src.mutation
}

// Save creates the SystemRelationship in the database.
func (src *SystemRelationshipCreate) Save(ctx context.Context) (*SystemRelationship, error) {
	src.defaults()
	return withHooks(ctx, src.sqlSave, src.mutation, src.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (src *SystemRelationshipCreate) SaveX(ctx context.Context) *SystemRelationship {
	v, err := src.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (src *SystemRelationshipCreate) Exec(ctx context.Context) error {
	_, err := src.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (src *SystemRelationshipCreate) ExecX(ctx context.Context) {
	if err := src.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (src *SystemRelationshipCreate) defaults() {
	if _, ok := src.mutation.CreatedAt(); !ok {
		v := systemrelationship.DefaultCreatedAt()
		src.mutation.SetCreatedAt(v)
	}
	if _, ok := src.mutation.ID(); !ok {
		v := systemrelationship.DefaultID()
		src.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (src *SystemRelationshipCreate) check() error {
	if _, ok := src.mutation.SourceComponentID(); !ok {
		return &ValidationError{Name: "source_component_id", err: errors.New(`ent: missing required field "SystemRelationship.source_component_id"`)}
	}
	if _, ok := src.mutation.TargetComponentID(); !ok {
		return &ValidationError{Name: "target_component_id", err: errors.New(`ent: missing required field "SystemRelationship.target_component_id"`)}
	}
	if _, ok := src.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "SystemRelationship.created_at"`)}
	}
	if len(src.mutation.SourceComponentIDs()) == 0 {
		return &ValidationError{Name: "source_component", err: errors.New(`ent: missing required edge "SystemRelationship.source_component"`)}
	}
	if len(src.mutation.TargetComponentIDs()) == 0 {
		return &ValidationError{Name: "target_component", err: errors.New(`ent: missing required edge "SystemRelationship.target_component"`)}
	}
	return nil
}

func (src *SystemRelationshipCreate) sqlSave(ctx context.Context) (*SystemRelationship, error) {
	if err := src.check(); err != nil {
		return nil, err
	}
	_node, _spec := src.createSpec()
	if err := sqlgraph.CreateNode(ctx, src.driver, _spec); err != nil {
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
	src.mutation.id = &_node.ID
	src.mutation.done = true
	return _node, nil
}

func (src *SystemRelationshipCreate) createSpec() (*SystemRelationship, *sqlgraph.CreateSpec) {
	var (
		_node = &SystemRelationship{config: src.config}
		_spec = sqlgraph.NewCreateSpec(systemrelationship.Table, sqlgraph.NewFieldSpec(systemrelationship.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = src.conflict
	if id, ok := src.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := src.mutation.Description(); ok {
		_spec.SetField(systemrelationship.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if value, ok := src.mutation.CreatedAt(); ok {
		_spec.SetField(systemrelationship.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if nodes := src.mutation.SourceComponentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationship.SourceComponentTable,
			Columns: []string{systemrelationship.SourceComponentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.SourceComponentID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := src.mutation.TargetComponentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationship.TargetComponentTable,
			Columns: []string{systemrelationship.TargetComponentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.TargetComponentID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := src.mutation.ControlsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemrelationship.ControlsTable,
			Columns: systemrelationship.ControlsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentcontrol.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		createE := &SystemRelationshipControlActionCreate{config: src.config, mutation: newSystemRelationshipControlActionMutation(src.config, OpCreate)}
		createE.defaults()
		_, specE := createE.createSpec()
		edge.Target.Fields = specE.Fields
		if specE.ID.Value != nil {
			edge.Target.Fields = append(edge.Target.Fields, specE.ID)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := src.mutation.SignalsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemrelationship.SignalsTable,
			Columns: systemrelationship.SignalsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentsignal.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		createE := &SystemRelationshipFeedbackCreate{config: src.config, mutation: newSystemRelationshipFeedbackMutation(src.config, OpCreate)}
		createE.defaults()
		_, specE := createE.createSpec()
		edge.Target.Fields = specE.Fields
		if specE.ID.Value != nil {
			edge.Target.Fields = append(edge.Target.Fields, specE.ID)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := src.mutation.ControlActionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemrelationship.ControlActionsTable,
			Columns: []string{systemrelationship.ControlActionsColumn},
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
	if nodes := src.mutation.FeedbackIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemrelationship.FeedbackTable,
			Columns: []string{systemrelationship.FeedbackColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemrelationshipfeedback.FieldID, field.TypeUUID),
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
//	client.SystemRelationship.Create().
//		SetSourceComponentID(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.SystemRelationshipUpsert) {
//			SetSourceComponentID(v+v).
//		}).
//		Exec(ctx)
func (src *SystemRelationshipCreate) OnConflict(opts ...sql.ConflictOption) *SystemRelationshipUpsertOne {
	src.conflict = opts
	return &SystemRelationshipUpsertOne{
		create: src,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.SystemRelationship.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (src *SystemRelationshipCreate) OnConflictColumns(columns ...string) *SystemRelationshipUpsertOne {
	src.conflict = append(src.conflict, sql.ConflictColumns(columns...))
	return &SystemRelationshipUpsertOne{
		create: src,
	}
}

type (
	// SystemRelationshipUpsertOne is the builder for "upsert"-ing
	//  one SystemRelationship node.
	SystemRelationshipUpsertOne struct {
		create *SystemRelationshipCreate
	}

	// SystemRelationshipUpsert is the "OnConflict" setter.
	SystemRelationshipUpsert struct {
		*sql.UpdateSet
	}
)

// SetSourceComponentID sets the "source_component_id" field.
func (u *SystemRelationshipUpsert) SetSourceComponentID(v uuid.UUID) *SystemRelationshipUpsert {
	u.Set(systemrelationship.FieldSourceComponentID, v)
	return u
}

// UpdateSourceComponentID sets the "source_component_id" field to the value that was provided on create.
func (u *SystemRelationshipUpsert) UpdateSourceComponentID() *SystemRelationshipUpsert {
	u.SetExcluded(systemrelationship.FieldSourceComponentID)
	return u
}

// SetTargetComponentID sets the "target_component_id" field.
func (u *SystemRelationshipUpsert) SetTargetComponentID(v uuid.UUID) *SystemRelationshipUpsert {
	u.Set(systemrelationship.FieldTargetComponentID, v)
	return u
}

// UpdateTargetComponentID sets the "target_component_id" field to the value that was provided on create.
func (u *SystemRelationshipUpsert) UpdateTargetComponentID() *SystemRelationshipUpsert {
	u.SetExcluded(systemrelationship.FieldTargetComponentID)
	return u
}

// SetDescription sets the "description" field.
func (u *SystemRelationshipUpsert) SetDescription(v string) *SystemRelationshipUpsert {
	u.Set(systemrelationship.FieldDescription, v)
	return u
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *SystemRelationshipUpsert) UpdateDescription() *SystemRelationshipUpsert {
	u.SetExcluded(systemrelationship.FieldDescription)
	return u
}

// ClearDescription clears the value of the "description" field.
func (u *SystemRelationshipUpsert) ClearDescription() *SystemRelationshipUpsert {
	u.SetNull(systemrelationship.FieldDescription)
	return u
}

// SetCreatedAt sets the "created_at" field.
func (u *SystemRelationshipUpsert) SetCreatedAt(v time.Time) *SystemRelationshipUpsert {
	u.Set(systemrelationship.FieldCreatedAt, v)
	return u
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *SystemRelationshipUpsert) UpdateCreatedAt() *SystemRelationshipUpsert {
	u.SetExcluded(systemrelationship.FieldCreatedAt)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.SystemRelationship.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(systemrelationship.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *SystemRelationshipUpsertOne) UpdateNewValues() *SystemRelationshipUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(systemrelationship.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.SystemRelationship.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *SystemRelationshipUpsertOne) Ignore() *SystemRelationshipUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *SystemRelationshipUpsertOne) DoNothing() *SystemRelationshipUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the SystemRelationshipCreate.OnConflict
// documentation for more info.
func (u *SystemRelationshipUpsertOne) Update(set func(*SystemRelationshipUpsert)) *SystemRelationshipUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&SystemRelationshipUpsert{UpdateSet: update})
	}))
	return u
}

// SetSourceComponentID sets the "source_component_id" field.
func (u *SystemRelationshipUpsertOne) SetSourceComponentID(v uuid.UUID) *SystemRelationshipUpsertOne {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.SetSourceComponentID(v)
	})
}

// UpdateSourceComponentID sets the "source_component_id" field to the value that was provided on create.
func (u *SystemRelationshipUpsertOne) UpdateSourceComponentID() *SystemRelationshipUpsertOne {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.UpdateSourceComponentID()
	})
}

// SetTargetComponentID sets the "target_component_id" field.
func (u *SystemRelationshipUpsertOne) SetTargetComponentID(v uuid.UUID) *SystemRelationshipUpsertOne {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.SetTargetComponentID(v)
	})
}

// UpdateTargetComponentID sets the "target_component_id" field to the value that was provided on create.
func (u *SystemRelationshipUpsertOne) UpdateTargetComponentID() *SystemRelationshipUpsertOne {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.UpdateTargetComponentID()
	})
}

// SetDescription sets the "description" field.
func (u *SystemRelationshipUpsertOne) SetDescription(v string) *SystemRelationshipUpsertOne {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *SystemRelationshipUpsertOne) UpdateDescription() *SystemRelationshipUpsertOne {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *SystemRelationshipUpsertOne) ClearDescription() *SystemRelationshipUpsertOne {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.ClearDescription()
	})
}

// SetCreatedAt sets the "created_at" field.
func (u *SystemRelationshipUpsertOne) SetCreatedAt(v time.Time) *SystemRelationshipUpsertOne {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.SetCreatedAt(v)
	})
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *SystemRelationshipUpsertOne) UpdateCreatedAt() *SystemRelationshipUpsertOne {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.UpdateCreatedAt()
	})
}

// Exec executes the query.
func (u *SystemRelationshipUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for SystemRelationshipCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *SystemRelationshipUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *SystemRelationshipUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: SystemRelationshipUpsertOne.ID is not supported by MySQL driver. Use SystemRelationshipUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *SystemRelationshipUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// SystemRelationshipCreateBulk is the builder for creating many SystemRelationship entities in bulk.
type SystemRelationshipCreateBulk struct {
	config
	err      error
	builders []*SystemRelationshipCreate
	conflict []sql.ConflictOption
}

// Save creates the SystemRelationship entities in the database.
func (srcb *SystemRelationshipCreateBulk) Save(ctx context.Context) ([]*SystemRelationship, error) {
	if srcb.err != nil {
		return nil, srcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(srcb.builders))
	nodes := make([]*SystemRelationship, len(srcb.builders))
	mutators := make([]Mutator, len(srcb.builders))
	for i := range srcb.builders {
		func(i int, root context.Context) {
			builder := srcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SystemRelationshipMutation)
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
					_, err = mutators[i+1].Mutate(root, srcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = srcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, srcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, srcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (srcb *SystemRelationshipCreateBulk) SaveX(ctx context.Context) []*SystemRelationship {
	v, err := srcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (srcb *SystemRelationshipCreateBulk) Exec(ctx context.Context) error {
	_, err := srcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (srcb *SystemRelationshipCreateBulk) ExecX(ctx context.Context) {
	if err := srcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.SystemRelationship.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.SystemRelationshipUpsert) {
//			SetSourceComponentID(v+v).
//		}).
//		Exec(ctx)
func (srcb *SystemRelationshipCreateBulk) OnConflict(opts ...sql.ConflictOption) *SystemRelationshipUpsertBulk {
	srcb.conflict = opts
	return &SystemRelationshipUpsertBulk{
		create: srcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.SystemRelationship.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (srcb *SystemRelationshipCreateBulk) OnConflictColumns(columns ...string) *SystemRelationshipUpsertBulk {
	srcb.conflict = append(srcb.conflict, sql.ConflictColumns(columns...))
	return &SystemRelationshipUpsertBulk{
		create: srcb,
	}
}

// SystemRelationshipUpsertBulk is the builder for "upsert"-ing
// a bulk of SystemRelationship nodes.
type SystemRelationshipUpsertBulk struct {
	create *SystemRelationshipCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.SystemRelationship.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(systemrelationship.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *SystemRelationshipUpsertBulk) UpdateNewValues() *SystemRelationshipUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(systemrelationship.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.SystemRelationship.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *SystemRelationshipUpsertBulk) Ignore() *SystemRelationshipUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *SystemRelationshipUpsertBulk) DoNothing() *SystemRelationshipUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the SystemRelationshipCreateBulk.OnConflict
// documentation for more info.
func (u *SystemRelationshipUpsertBulk) Update(set func(*SystemRelationshipUpsert)) *SystemRelationshipUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&SystemRelationshipUpsert{UpdateSet: update})
	}))
	return u
}

// SetSourceComponentID sets the "source_component_id" field.
func (u *SystemRelationshipUpsertBulk) SetSourceComponentID(v uuid.UUID) *SystemRelationshipUpsertBulk {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.SetSourceComponentID(v)
	})
}

// UpdateSourceComponentID sets the "source_component_id" field to the value that was provided on create.
func (u *SystemRelationshipUpsertBulk) UpdateSourceComponentID() *SystemRelationshipUpsertBulk {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.UpdateSourceComponentID()
	})
}

// SetTargetComponentID sets the "target_component_id" field.
func (u *SystemRelationshipUpsertBulk) SetTargetComponentID(v uuid.UUID) *SystemRelationshipUpsertBulk {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.SetTargetComponentID(v)
	})
}

// UpdateTargetComponentID sets the "target_component_id" field to the value that was provided on create.
func (u *SystemRelationshipUpsertBulk) UpdateTargetComponentID() *SystemRelationshipUpsertBulk {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.UpdateTargetComponentID()
	})
}

// SetDescription sets the "description" field.
func (u *SystemRelationshipUpsertBulk) SetDescription(v string) *SystemRelationshipUpsertBulk {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *SystemRelationshipUpsertBulk) UpdateDescription() *SystemRelationshipUpsertBulk {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *SystemRelationshipUpsertBulk) ClearDescription() *SystemRelationshipUpsertBulk {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.ClearDescription()
	})
}

// SetCreatedAt sets the "created_at" field.
func (u *SystemRelationshipUpsertBulk) SetCreatedAt(v time.Time) *SystemRelationshipUpsertBulk {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.SetCreatedAt(v)
	})
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *SystemRelationshipUpsertBulk) UpdateCreatedAt() *SystemRelationshipUpsertBulk {
	return u.Update(func(s *SystemRelationshipUpsert) {
		s.UpdateCreatedAt()
	})
}

// Exec executes the query.
func (u *SystemRelationshipUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the SystemRelationshipCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for SystemRelationshipCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *SystemRelationshipUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
