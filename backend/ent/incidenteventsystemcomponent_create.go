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
	"github.com/rezible/rezible/ent/incidenteventsystemcomponent"
	"github.com/rezible/rezible/ent/systemcomponent"
)

// IncidentEventSystemComponentCreate is the builder for creating a IncidentEventSystemComponent entity.
type IncidentEventSystemComponentCreate struct {
	config
	mutation *IncidentEventSystemComponentMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetIncidentEventID sets the "incident_event_id" field.
func (iescc *IncidentEventSystemComponentCreate) SetIncidentEventID(u uuid.UUID) *IncidentEventSystemComponentCreate {
	iescc.mutation.SetIncidentEventID(u)
	return iescc
}

// SetSystemComponentID sets the "system_component_id" field.
func (iescc *IncidentEventSystemComponentCreate) SetSystemComponentID(u uuid.UUID) *IncidentEventSystemComponentCreate {
	iescc.mutation.SetSystemComponentID(u)
	return iescc
}

// SetRelationship sets the "relationship" field.
func (iescc *IncidentEventSystemComponentCreate) SetRelationship(i incidenteventsystemcomponent.Relationship) *IncidentEventSystemComponentCreate {
	iescc.mutation.SetRelationship(i)
	return iescc
}

// SetCreatedAt sets the "created_at" field.
func (iescc *IncidentEventSystemComponentCreate) SetCreatedAt(t time.Time) *IncidentEventSystemComponentCreate {
	iescc.mutation.SetCreatedAt(t)
	return iescc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (iescc *IncidentEventSystemComponentCreate) SetNillableCreatedAt(t *time.Time) *IncidentEventSystemComponentCreate {
	if t != nil {
		iescc.SetCreatedAt(*t)
	}
	return iescc
}

// SetID sets the "id" field.
func (iescc *IncidentEventSystemComponentCreate) SetID(u uuid.UUID) *IncidentEventSystemComponentCreate {
	iescc.mutation.SetID(u)
	return iescc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (iescc *IncidentEventSystemComponentCreate) SetNillableID(u *uuid.UUID) *IncidentEventSystemComponentCreate {
	if u != nil {
		iescc.SetID(*u)
	}
	return iescc
}

// SetEventID sets the "event" edge to the IncidentEventSystemComponent entity by ID.
func (iescc *IncidentEventSystemComponentCreate) SetEventID(id uuid.UUID) *IncidentEventSystemComponentCreate {
	iescc.mutation.SetEventID(id)
	return iescc
}

// SetEvent sets the "event" edge to the IncidentEventSystemComponent entity.
func (iescc *IncidentEventSystemComponentCreate) SetEvent(i *IncidentEventSystemComponent) *IncidentEventSystemComponentCreate {
	return iescc.SetEventID(i.ID)
}

// SetSystemComponent sets the "system_component" edge to the SystemComponent entity.
func (iescc *IncidentEventSystemComponentCreate) SetSystemComponent(s *SystemComponent) *IncidentEventSystemComponentCreate {
	return iescc.SetSystemComponentID(s.ID)
}

// Mutation returns the IncidentEventSystemComponentMutation object of the builder.
func (iescc *IncidentEventSystemComponentCreate) Mutation() *IncidentEventSystemComponentMutation {
	return iescc.mutation
}

// Save creates the IncidentEventSystemComponent in the database.
func (iescc *IncidentEventSystemComponentCreate) Save(ctx context.Context) (*IncidentEventSystemComponent, error) {
	iescc.defaults()
	return withHooks(ctx, iescc.sqlSave, iescc.mutation, iescc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (iescc *IncidentEventSystemComponentCreate) SaveX(ctx context.Context) *IncidentEventSystemComponent {
	v, err := iescc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (iescc *IncidentEventSystemComponentCreate) Exec(ctx context.Context) error {
	_, err := iescc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (iescc *IncidentEventSystemComponentCreate) ExecX(ctx context.Context) {
	if err := iescc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (iescc *IncidentEventSystemComponentCreate) defaults() {
	if _, ok := iescc.mutation.CreatedAt(); !ok {
		v := incidenteventsystemcomponent.DefaultCreatedAt()
		iescc.mutation.SetCreatedAt(v)
	}
	if _, ok := iescc.mutation.ID(); !ok {
		v := incidenteventsystemcomponent.DefaultID()
		iescc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (iescc *IncidentEventSystemComponentCreate) check() error {
	if _, ok := iescc.mutation.IncidentEventID(); !ok {
		return &ValidationError{Name: "incident_event_id", err: errors.New(`ent: missing required field "IncidentEventSystemComponent.incident_event_id"`)}
	}
	if _, ok := iescc.mutation.SystemComponentID(); !ok {
		return &ValidationError{Name: "system_component_id", err: errors.New(`ent: missing required field "IncidentEventSystemComponent.system_component_id"`)}
	}
	if _, ok := iescc.mutation.Relationship(); !ok {
		return &ValidationError{Name: "relationship", err: errors.New(`ent: missing required field "IncidentEventSystemComponent.relationship"`)}
	}
	if v, ok := iescc.mutation.Relationship(); ok {
		if err := incidenteventsystemcomponent.RelationshipValidator(v); err != nil {
			return &ValidationError{Name: "relationship", err: fmt.Errorf(`ent: validator failed for field "IncidentEventSystemComponent.relationship": %w`, err)}
		}
	}
	if _, ok := iescc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "IncidentEventSystemComponent.created_at"`)}
	}
	if len(iescc.mutation.EventIDs()) == 0 {
		return &ValidationError{Name: "event", err: errors.New(`ent: missing required edge "IncidentEventSystemComponent.event"`)}
	}
	if len(iescc.mutation.SystemComponentIDs()) == 0 {
		return &ValidationError{Name: "system_component", err: errors.New(`ent: missing required edge "IncidentEventSystemComponent.system_component"`)}
	}
	return nil
}

func (iescc *IncidentEventSystemComponentCreate) sqlSave(ctx context.Context) (*IncidentEventSystemComponent, error) {
	if err := iescc.check(); err != nil {
		return nil, err
	}
	_node, _spec := iescc.createSpec()
	if err := sqlgraph.CreateNode(ctx, iescc.driver, _spec); err != nil {
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
	iescc.mutation.id = &_node.ID
	iescc.mutation.done = true
	return _node, nil
}

func (iescc *IncidentEventSystemComponentCreate) createSpec() (*IncidentEventSystemComponent, *sqlgraph.CreateSpec) {
	var (
		_node = &IncidentEventSystemComponent{config: iescc.config}
		_spec = sqlgraph.NewCreateSpec(incidenteventsystemcomponent.Table, sqlgraph.NewFieldSpec(incidenteventsystemcomponent.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = iescc.conflict
	if id, ok := iescc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := iescc.mutation.Relationship(); ok {
		_spec.SetField(incidenteventsystemcomponent.FieldRelationship, field.TypeEnum, value)
		_node.Relationship = value
	}
	if value, ok := iescc.mutation.CreatedAt(); ok {
		_spec.SetField(incidenteventsystemcomponent.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if nodes := iescc.mutation.EventIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   incidenteventsystemcomponent.EventTable,
			Columns: []string{incidenteventsystemcomponent.EventColumn},
			Bidi:    true,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidenteventsystemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.IncidentEventID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := iescc.mutation.SystemComponentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   incidenteventsystemcomponent.SystemComponentTable,
			Columns: []string{incidenteventsystemcomponent.SystemComponentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.SystemComponentID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.IncidentEventSystemComponent.Create().
//		SetIncidentEventID(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentEventSystemComponentUpsert) {
//			SetIncidentEventID(v+v).
//		}).
//		Exec(ctx)
func (iescc *IncidentEventSystemComponentCreate) OnConflict(opts ...sql.ConflictOption) *IncidentEventSystemComponentUpsertOne {
	iescc.conflict = opts
	return &IncidentEventSystemComponentUpsertOne{
		create: iescc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentEventSystemComponent.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (iescc *IncidentEventSystemComponentCreate) OnConflictColumns(columns ...string) *IncidentEventSystemComponentUpsertOne {
	iescc.conflict = append(iescc.conflict, sql.ConflictColumns(columns...))
	return &IncidentEventSystemComponentUpsertOne{
		create: iescc,
	}
}

type (
	// IncidentEventSystemComponentUpsertOne is the builder for "upsert"-ing
	//  one IncidentEventSystemComponent node.
	IncidentEventSystemComponentUpsertOne struct {
		create *IncidentEventSystemComponentCreate
	}

	// IncidentEventSystemComponentUpsert is the "OnConflict" setter.
	IncidentEventSystemComponentUpsert struct {
		*sql.UpdateSet
	}
)

// SetIncidentEventID sets the "incident_event_id" field.
func (u *IncidentEventSystemComponentUpsert) SetIncidentEventID(v uuid.UUID) *IncidentEventSystemComponentUpsert {
	u.Set(incidenteventsystemcomponent.FieldIncidentEventID, v)
	return u
}

// UpdateIncidentEventID sets the "incident_event_id" field to the value that was provided on create.
func (u *IncidentEventSystemComponentUpsert) UpdateIncidentEventID() *IncidentEventSystemComponentUpsert {
	u.SetExcluded(incidenteventsystemcomponent.FieldIncidentEventID)
	return u
}

// SetSystemComponentID sets the "system_component_id" field.
func (u *IncidentEventSystemComponentUpsert) SetSystemComponentID(v uuid.UUID) *IncidentEventSystemComponentUpsert {
	u.Set(incidenteventsystemcomponent.FieldSystemComponentID, v)
	return u
}

// UpdateSystemComponentID sets the "system_component_id" field to the value that was provided on create.
func (u *IncidentEventSystemComponentUpsert) UpdateSystemComponentID() *IncidentEventSystemComponentUpsert {
	u.SetExcluded(incidenteventsystemcomponent.FieldSystemComponentID)
	return u
}

// SetRelationship sets the "relationship" field.
func (u *IncidentEventSystemComponentUpsert) SetRelationship(v incidenteventsystemcomponent.Relationship) *IncidentEventSystemComponentUpsert {
	u.Set(incidenteventsystemcomponent.FieldRelationship, v)
	return u
}

// UpdateRelationship sets the "relationship" field to the value that was provided on create.
func (u *IncidentEventSystemComponentUpsert) UpdateRelationship() *IncidentEventSystemComponentUpsert {
	u.SetExcluded(incidenteventsystemcomponent.FieldRelationship)
	return u
}

// SetCreatedAt sets the "created_at" field.
func (u *IncidentEventSystemComponentUpsert) SetCreatedAt(v time.Time) *IncidentEventSystemComponentUpsert {
	u.Set(incidenteventsystemcomponent.FieldCreatedAt, v)
	return u
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *IncidentEventSystemComponentUpsert) UpdateCreatedAt() *IncidentEventSystemComponentUpsert {
	u.SetExcluded(incidenteventsystemcomponent.FieldCreatedAt)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.IncidentEventSystemComponent.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidenteventsystemcomponent.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentEventSystemComponentUpsertOne) UpdateNewValues() *IncidentEventSystemComponentUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(incidenteventsystemcomponent.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentEventSystemComponent.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *IncidentEventSystemComponentUpsertOne) Ignore() *IncidentEventSystemComponentUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentEventSystemComponentUpsertOne) DoNothing() *IncidentEventSystemComponentUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentEventSystemComponentCreate.OnConflict
// documentation for more info.
func (u *IncidentEventSystemComponentUpsertOne) Update(set func(*IncidentEventSystemComponentUpsert)) *IncidentEventSystemComponentUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentEventSystemComponentUpsert{UpdateSet: update})
	}))
	return u
}

// SetIncidentEventID sets the "incident_event_id" field.
func (u *IncidentEventSystemComponentUpsertOne) SetIncidentEventID(v uuid.UUID) *IncidentEventSystemComponentUpsertOne {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.SetIncidentEventID(v)
	})
}

// UpdateIncidentEventID sets the "incident_event_id" field to the value that was provided on create.
func (u *IncidentEventSystemComponentUpsertOne) UpdateIncidentEventID() *IncidentEventSystemComponentUpsertOne {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.UpdateIncidentEventID()
	})
}

// SetSystemComponentID sets the "system_component_id" field.
func (u *IncidentEventSystemComponentUpsertOne) SetSystemComponentID(v uuid.UUID) *IncidentEventSystemComponentUpsertOne {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.SetSystemComponentID(v)
	})
}

// UpdateSystemComponentID sets the "system_component_id" field to the value that was provided on create.
func (u *IncidentEventSystemComponentUpsertOne) UpdateSystemComponentID() *IncidentEventSystemComponentUpsertOne {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.UpdateSystemComponentID()
	})
}

// SetRelationship sets the "relationship" field.
func (u *IncidentEventSystemComponentUpsertOne) SetRelationship(v incidenteventsystemcomponent.Relationship) *IncidentEventSystemComponentUpsertOne {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.SetRelationship(v)
	})
}

// UpdateRelationship sets the "relationship" field to the value that was provided on create.
func (u *IncidentEventSystemComponentUpsertOne) UpdateRelationship() *IncidentEventSystemComponentUpsertOne {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.UpdateRelationship()
	})
}

// SetCreatedAt sets the "created_at" field.
func (u *IncidentEventSystemComponentUpsertOne) SetCreatedAt(v time.Time) *IncidentEventSystemComponentUpsertOne {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.SetCreatedAt(v)
	})
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *IncidentEventSystemComponentUpsertOne) UpdateCreatedAt() *IncidentEventSystemComponentUpsertOne {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.UpdateCreatedAt()
	})
}

// Exec executes the query.
func (u *IncidentEventSystemComponentUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentEventSystemComponentCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentEventSystemComponentUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *IncidentEventSystemComponentUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: IncidentEventSystemComponentUpsertOne.ID is not supported by MySQL driver. Use IncidentEventSystemComponentUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *IncidentEventSystemComponentUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// IncidentEventSystemComponentCreateBulk is the builder for creating many IncidentEventSystemComponent entities in bulk.
type IncidentEventSystemComponentCreateBulk struct {
	config
	err      error
	builders []*IncidentEventSystemComponentCreate
	conflict []sql.ConflictOption
}

// Save creates the IncidentEventSystemComponent entities in the database.
func (iesccb *IncidentEventSystemComponentCreateBulk) Save(ctx context.Context) ([]*IncidentEventSystemComponent, error) {
	if iesccb.err != nil {
		return nil, iesccb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(iesccb.builders))
	nodes := make([]*IncidentEventSystemComponent, len(iesccb.builders))
	mutators := make([]Mutator, len(iesccb.builders))
	for i := range iesccb.builders {
		func(i int, root context.Context) {
			builder := iesccb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*IncidentEventSystemComponentMutation)
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
					_, err = mutators[i+1].Mutate(root, iesccb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = iesccb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, iesccb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, iesccb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (iesccb *IncidentEventSystemComponentCreateBulk) SaveX(ctx context.Context) []*IncidentEventSystemComponent {
	v, err := iesccb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (iesccb *IncidentEventSystemComponentCreateBulk) Exec(ctx context.Context) error {
	_, err := iesccb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (iesccb *IncidentEventSystemComponentCreateBulk) ExecX(ctx context.Context) {
	if err := iesccb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.IncidentEventSystemComponent.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentEventSystemComponentUpsert) {
//			SetIncidentEventID(v+v).
//		}).
//		Exec(ctx)
func (iesccb *IncidentEventSystemComponentCreateBulk) OnConflict(opts ...sql.ConflictOption) *IncidentEventSystemComponentUpsertBulk {
	iesccb.conflict = opts
	return &IncidentEventSystemComponentUpsertBulk{
		create: iesccb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentEventSystemComponent.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (iesccb *IncidentEventSystemComponentCreateBulk) OnConflictColumns(columns ...string) *IncidentEventSystemComponentUpsertBulk {
	iesccb.conflict = append(iesccb.conflict, sql.ConflictColumns(columns...))
	return &IncidentEventSystemComponentUpsertBulk{
		create: iesccb,
	}
}

// IncidentEventSystemComponentUpsertBulk is the builder for "upsert"-ing
// a bulk of IncidentEventSystemComponent nodes.
type IncidentEventSystemComponentUpsertBulk struct {
	create *IncidentEventSystemComponentCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.IncidentEventSystemComponent.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidenteventsystemcomponent.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentEventSystemComponentUpsertBulk) UpdateNewValues() *IncidentEventSystemComponentUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(incidenteventsystemcomponent.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentEventSystemComponent.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *IncidentEventSystemComponentUpsertBulk) Ignore() *IncidentEventSystemComponentUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentEventSystemComponentUpsertBulk) DoNothing() *IncidentEventSystemComponentUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentEventSystemComponentCreateBulk.OnConflict
// documentation for more info.
func (u *IncidentEventSystemComponentUpsertBulk) Update(set func(*IncidentEventSystemComponentUpsert)) *IncidentEventSystemComponentUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentEventSystemComponentUpsert{UpdateSet: update})
	}))
	return u
}

// SetIncidentEventID sets the "incident_event_id" field.
func (u *IncidentEventSystemComponentUpsertBulk) SetIncidentEventID(v uuid.UUID) *IncidentEventSystemComponentUpsertBulk {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.SetIncidentEventID(v)
	})
}

// UpdateIncidentEventID sets the "incident_event_id" field to the value that was provided on create.
func (u *IncidentEventSystemComponentUpsertBulk) UpdateIncidentEventID() *IncidentEventSystemComponentUpsertBulk {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.UpdateIncidentEventID()
	})
}

// SetSystemComponentID sets the "system_component_id" field.
func (u *IncidentEventSystemComponentUpsertBulk) SetSystemComponentID(v uuid.UUID) *IncidentEventSystemComponentUpsertBulk {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.SetSystemComponentID(v)
	})
}

// UpdateSystemComponentID sets the "system_component_id" field to the value that was provided on create.
func (u *IncidentEventSystemComponentUpsertBulk) UpdateSystemComponentID() *IncidentEventSystemComponentUpsertBulk {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.UpdateSystemComponentID()
	})
}

// SetRelationship sets the "relationship" field.
func (u *IncidentEventSystemComponentUpsertBulk) SetRelationship(v incidenteventsystemcomponent.Relationship) *IncidentEventSystemComponentUpsertBulk {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.SetRelationship(v)
	})
}

// UpdateRelationship sets the "relationship" field to the value that was provided on create.
func (u *IncidentEventSystemComponentUpsertBulk) UpdateRelationship() *IncidentEventSystemComponentUpsertBulk {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.UpdateRelationship()
	})
}

// SetCreatedAt sets the "created_at" field.
func (u *IncidentEventSystemComponentUpsertBulk) SetCreatedAt(v time.Time) *IncidentEventSystemComponentUpsertBulk {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.SetCreatedAt(v)
	})
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *IncidentEventSystemComponentUpsertBulk) UpdateCreatedAt() *IncidentEventSystemComponentUpsertBulk {
	return u.Update(func(s *IncidentEventSystemComponentUpsert) {
		s.UpdateCreatedAt()
	})
}

// Exec executes the query.
func (u *IncidentEventSystemComponentUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the IncidentEventSystemComponentCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentEventSystemComponentCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentEventSystemComponentUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
